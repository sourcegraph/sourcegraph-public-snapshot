import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import { parseSearchURLQuery, parseSearchURLPatternType, PatternTypeProps, InteractiveSearchProps } from '..'
import { Contributions, Evaluated } from '../../../../shared/src/api/protocol'
import { FetchFileCtx } from '../../../../shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { PageTitle } from '../../components/PageTitle'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { EventLogger } from '../../tracking/eventLogger'
import { isSearchResults, submitSearch, toggleSearchFilter, getSearchTypeFromQuery, QueryState } from '../helpers'
import { queryTelemetryData } from '../queryTelemetry'
import { SearchResultsFilterBars, SearchScopeWithOptionalName } from './SearchResultsFilterBars'
import { SearchResultsList } from './SearchResultsList'
import { SearchResultTypeTabs } from './SearchResultTypeTabs'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'

export interface SearchResultsProps
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        SettingsCascadeProps,
        TelemetryProps,
        ThemeProps,
        PatternTypeProps,
        InteractiveSearchProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    navbarSearchQueryState: QueryState
    telemetryService: Pick<EventLogger, 'log' | 'logViewEvent'>
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
    searchRequest: (
        query: string,
        version: string,
        patternType: GQL.SearchPatternType,
        { extensionsController }: ExtensionsControllerProps<'services'>
    ) => Observable<GQL.ISearchResults | ErrorLike>
    isSourcegraphDotCom: boolean
    deployType: DeployType
    filtersInQuery: FiltersToTypeAndValue
    interactiveSearchMode: boolean
}

interface SearchResultsState {
    /** The loaded search results, error or undefined while loading */
    resultsOrError?: GQL.ISearchResults
    allExpanded: boolean

    // Saved Queries
    showSavedQueryModal: boolean
    didSaveQuery: boolean

    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Evaluated<Contributions>
}

/** All values that are valid for the `type:` filter. `null` represents default code search. */
export type SearchType = 'diff' | 'commit' | 'symbol' | 'repo' | 'path' | null

// The latest supported version of our search syntax. Users should never be able to determine the search version.
// The version is set based on the release tag of the instance. Anything before 3.9.0 will not pass a version parameter,
// and will therefore default to V1.
const LATEST_VERSION = 'V2'

export class SearchResults extends React.Component<SearchResultsProps, SearchResultsState> {
    public state: SearchResultsState = {
        didSaveQuery: false,
        showSavedQueryModal: false,
        allExpanded: false,
    }
    /** Emits on componentDidUpdate with the new props */
    private componentUpdates = new Subject<SearchResultsProps>()

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const patternType = parseSearchURLPatternType(that.props.location.search)

        if (!patternType) {
            // If the patternType query parameter does not exist in the URL or is invalid, redirect to a URL which
            // has patternType=regexp appended. This is to ensure old URLs before requiring patternType still work.
            const newLoc =
                '/search?' +
                buildSearchURLQuery(
                    that.props.navbarSearchQueryState.query,
                    GQL.SearchPatternType.regexp,
                    that.props.filtersInQuery
                )
            window.location.replace(newLoc)
        }

        that.props.telemetryService.logViewEvent('SearchResults')

        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    startWith(that.props),
                    map(props => [
                        parseSearchURLQuery(props.location.search, props.interactiveSearchMode),
                        parseSearchURLPatternType(props.location.search),
                    ]),
                    // Search when a new search query was specified in the URL
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    filter(
                        (queryAndPatternType): queryAndPatternType is [string, GQL.SearchPatternType] =>
                            !!queryAndPatternType[0] && !!queryAndPatternType[1]
                    ),
                    tap(([query]) => {
                        const query_data = queryTelemetryData(query)
                        that.props.telemetryService.log('SearchResultsQueried', {
                            code_search: { query_data },
                        })
                        if (
                            query_data.query &&
                            query_data.query.field_type &&
                            query_data.query.field_type.value_diff > 0
                        ) {
                            that.props.telemetryService.log('DiffSearchResultsQueried')
                        }
                    }),
                    switchMap(([query, patternType]) =>
                        concat(
                            // Reset view state
                            [
                                {
                                    resultsOrError: undefined,
                                    didSave: false,
                                    activeType: getSearchTypeFromQuery(query),
                                },
                            ],
                            // Do async search request
                            that.props.searchRequest(query, LATEST_VERSION, patternType, that.props).pipe(
                                // Log telemetry
                                tap(
                                    results => {
                                        that.props.telemetryService.log('SearchResultsFetched', {
                                            code_search: {
                                                // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                                results: {
                                                    results_count: isErrorLike(results) ? 0 : results.results.length,
                                                    any_cloning: isErrorLike(results)
                                                        ? false
                                                        : results.cloning.length > 0,
                                                },
                                            },
                                        })
                                        if (patternType && patternType !== that.props.patternType) {
                                            that.props.setPatternType(patternType)
                                        }
                                    },
                                    error => {
                                        that.props.telemetryService.log('SearchResultsFetchFailed', {
                                            code_search: { error_message: error.message },
                                        })
                                        console.error(error)
                                    }
                                ),
                                // Update view with results or error
                                map(resultsOrError => ({ resultsOrError })),
                                catchError(error => [{ resultsOrError: error }])
                            )
                        )
                    )
                )
                .subscribe(
                    newState => that.setState(newState as SearchResultsState),
                    err => console.error(err)
                )
        )

        that.props.extensionsController.services.contribution
            .getContributions()
            .subscribe(contributions => that.setState({ contributions }))
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    private showSaveQueryModal = (): void => {
        that.setState({ showSavedQueryModal: true, didSaveQuery: false })
    }

    private onDidCreateSavedQuery = (): void => {
        that.props.telemetryService.log('SavedQueryCreated')
        that.setState({ showSavedQueryModal: false, didSaveQuery: true })
    }

    private onModalClose = (): void => {
        that.props.telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        that.setState({ didSaveQuery: false, showSavedQueryModal: false })
    }

    public render(): JSX.Element | null {
        const query = parseSearchURLQuery(that.props.location.search, that.props.interactiveSearchMode)
        const filters = that.getFilters()
        const extensionFilters = that.state.contributions && that.state.contributions.searchFilters

        const quickLinks =
            (isSettingsValid<Settings>(that.props.settingsCascade) && that.props.settingsCascade.final.quicklinks) || []

        return (
            <div className="e2e-search-results search-results d-flex flex-column w-100">
                <PageTitle key="page-title" title={query} />
                {!that.props.interactiveSearchMode && (
                    <SearchResultsFilterBars
                        navbarSearchQuery={that.props.navbarSearchQueryState.query}
                        results={that.state.resultsOrError}
                        filters={filters}
                        extensionFilters={extensionFilters}
                        quickLinks={quickLinks}
                        onFilterClick={that.onDynamicFilterClicked}
                        onShowMoreResultsClick={that.showMoreResults}
                        calculateShowMoreResultsCount={that.calculateCount}
                    />
                )}
                <SearchResultTypeTabs
                    {...that.props}
                    query={that.props.navbarSearchQueryState.query}
                    filtersInQuery={that.props.filtersInQuery}
                />
                <SearchResultsList
                    {...that.props}
                    resultsOrError={that.state.resultsOrError}
                    onShowMoreResultsClick={that.showMoreResults}
                    onExpandAllResultsToggle={that.onExpandAllResultsToggle}
                    allExpanded={that.state.allExpanded}
                    showSavedQueryModal={that.state.showSavedQueryModal}
                    onSaveQueryClick={that.showSaveQueryModal}
                    onSavedQueryModalClose={that.onModalClose}
                    onDidCreateSavedQuery={that.onDidCreateSavedQuery}
                    didSave={that.state.didSaveQuery}
                />
            </div>
        )
    }

    /** Combines dynamic filters and search scopes into a list de-duplicated by value. */
    private getFilters(): SearchScopeWithOptionalName[] {
        const filters = new Map<string, SearchScopeWithOptionalName>()

        if (isSearchResults(that.state.resultsOrError) && that.state.resultsOrError.dynamicFilters) {
            let dynamicFilters = that.state.resultsOrError.dynamicFilters
            dynamicFilters = that.state.resultsOrError.dynamicFilters.filter(filter => filter.kind !== 'repo')
            for (const d of dynamicFilters) {
                filters.set(d.value, d)
            }
        }
        const scopes =
            (isSettingsValid<Settings>(that.props.settingsCascade) &&
                that.props.settingsCascade.final['search.scopes']) ||
            []
        if (isSearchResults(that.state.resultsOrError) && that.state.resultsOrError.dynamicFilters) {
            for (const scope of scopes) {
                if (!filters.has(scope.value)) {
                    filters.set(scope.value, scope)
                }
            }
        } else {
            for (const scope of scopes) {
                // Check for if filter.value already exists and if so, overwrite with user's configured scope name.
                const existingFilter = filters.get(scope.value)
                // This works because user setting configs are the last to be processed after Global and Org.
                // Thus, user set filters overwrite the equal valued existing filters.
                if (existingFilter) {
                    existingFilter.name = scope.name || scope.value
                }
                filters.set(scope.value, existingFilter || scope)
            }
        }

        return Array.from(filters.values())
    }
    private showMoreResults = (): void => {
        // Requery with an increased max result count.
        const params = new URLSearchParams(this.props.location.search)
        let query = params.get('q') || ''

        const count = this.calculateCount()
        if (/count:(\d+)/.test(query)) {
            query = query.replace(/count:\d+/g, '').trim() + ` count:${count}`
        } else {
            query = `${query} count:${count}`
        }
        params.set('q', query)
        this.props.history.replace({ search: params.toString() })
    }

    private calculateCount = (): number => {
        // This function can only get called if the results were successfully loaded,
        // so casting is the right thing to do here
        const results = this.state.resultsOrError as GQL.ISearchResults

        const params = new URLSearchParams(this.props.location.search)
        const query = params.get('q') || ''

        if (/count:(\d+)/.test(query)) {
            return Math.max(results.matchCount * 2, 1000)
        }
        return Math.max(results.matchCount * 2 || 0, 1000)
    }

    private onExpandAllResultsToggle = (): void => {
        this.setState(
            state => ({ allExpanded: !state.allExpanded }),
            () => {
                this.props.telemetryService.log(this.state.allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
            }
        )
    }

    private onDynamicFilterClicked = (value: string): void => {
        this.props.telemetryService.log('DynamicFilterClicked', {
            search_filter: { value },
        })

        const newQuery = toggleSearchFilter(this.props.navbarSearchQueryState.query, value)

        submitSearch(this.props.history, newQuery, 'filter', this.props.patternType)
    }
}
