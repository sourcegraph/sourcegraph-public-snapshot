import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import {
    parseSearchURLQuery,
    parseSearchURLPatternType,
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    parseSearchURL,
} from '..'
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
import { convertPlainTextToInteractiveQuery } from '../input/helpers'

export interface SearchResultsProps
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        SettingsCascadeProps,
        TelemetryProps,
        ThemeProps,
        PatternTypeProps,
        CaseSensitivityProps,
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
        const patternType = parseSearchURLPatternType(this.props.location.search)

        if (!patternType) {
            // If the patternType query parameter does not exist in the URL or is invalid, redirect to a URL which
            // has patternType=regexp appended. This is to ensure old URLs before requiring patternType still work.

            const q = parseSearchURLQuery(this.props.location.search) || ''
            const { navbarQuery, filtersInQuery } = convertPlainTextToInteractiveQuery(q)
            const newLoc =
                '/search?' +
                buildSearchURLQuery(navbarQuery, GQL.SearchPatternType.regexp, this.props.caseSensitive, filtersInQuery)
            window.location.replace(newLoc)
        }

        this.props.telemetryService.logViewEvent('SearchResults')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => parseSearchURL(props.location.search)),
                    // Search when a new search query was specified in the URL
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    filter(
                        (
                            queryAndPatternTypeAndCase
                        ): queryAndPatternTypeAndCase is {
                            query: string
                            patternType: GQL.SearchPatternType
                            caseSensitive: boolean
                        } => !!queryAndPatternTypeAndCase.query && !!queryAndPatternTypeAndCase.patternType
                    ),
                    tap(({ query, caseSensitive }) => {
                        const query_data = queryTelemetryData(query, caseSensitive)
                        this.props.telemetryService.log('SearchResultsQueried', {
                            code_search: { query_data },
                            ...(this.props.splitSearchModes
                                ? { mode: this.props.interactiveSearchMode ? 'interactive' : 'plain' }
                                : {}),
                        })
                        if (
                            query_data.query &&
                            query_data.query.field_type &&
                            query_data.query.field_type.value_diff > 0
                        ) {
                            this.props.telemetryService.log('DiffSearchResultsQueried')
                        }
                    }),
                    switchMap(({ query, patternType, caseSensitive }) =>
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
                            this.props
                                .searchRequest(
                                    caseSensitive ? `${query} case:yes` : query,
                                    LATEST_VERSION,
                                    patternType,
                                    this.props
                                )
                                .pipe(
                                    // Log telemetry
                                    tap(
                                        results => {
                                            this.props.telemetryService.log('SearchResultsFetched', {
                                                code_search: {
                                                    // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                                    results: {
                                                        results_count: isErrorLike(results)
                                                            ? 0
                                                            : results.results.length,
                                                        any_cloning: isErrorLike(results)
                                                            ? false
                                                            : results.cloning.length > 0,
                                                    },
                                                },
                                            })
                                            if (patternType && patternType !== this.props.patternType) {
                                                this.props.setPatternType(patternType)
                                            }
                                            if (caseSensitive !== this.props.caseSensitive) {
                                                this.props.setCaseSensitivity(caseSensitive)
                                            }
                                        },
                                        error => {
                                            this.props.telemetryService.log('SearchResultsFetchFailed', {
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
                    newState => this.setState(newState as SearchResultsState),
                    err => console.error(err)
                )
        )

        this.props.extensionsController.services.contribution
            .getContributions()
            .subscribe(contributions => this.setState({ contributions }))
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private showSaveQueryModal = (): void => {
        this.setState({ showSavedQueryModal: true, didSaveQuery: false })
    }

    private onDidCreateSavedQuery = (): void => {
        this.props.telemetryService.log('SavedQueryCreated')
        this.setState({ showSavedQueryModal: false, didSaveQuery: true })
    }

    private onModalClose = (): void => {
        this.props.telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        this.setState({ didSaveQuery: false, showSavedQueryModal: false })
    }

    public render(): JSX.Element | null {
        const query = parseSearchURLQuery(this.props.location.search)
        const filters = this.getFilters()
        const extensionFilters = this.state.contributions && this.state.contributions.searchFilters

        const quickLinks =
            (isSettingsValid<Settings>(this.props.settingsCascade) && this.props.settingsCascade.final.quicklinks) || []

        return (
            <div className="e2e-search-results search-results d-flex flex-column w-100">
                <PageTitle key="page-title" title={query} />
                {!this.props.interactiveSearchMode && (
                    <SearchResultsFilterBars
                        navbarSearchQuery={this.props.navbarSearchQueryState.query}
                        results={this.state.resultsOrError}
                        filters={filters}
                        extensionFilters={extensionFilters}
                        quickLinks={quickLinks}
                        onFilterClick={this.onDynamicFilterClicked}
                        onShowMoreResultsClick={this.showMoreResults}
                        calculateShowMoreResultsCount={this.calculateCount}
                    />
                )}
                <SearchResultTypeTabs
                    {...this.props}
                    query={this.props.navbarSearchQueryState.query}
                    filtersInQuery={this.props.filtersInQuery}
                />
                <SearchResultsList
                    {...this.props}
                    resultsOrError={this.state.resultsOrError}
                    onShowMoreResultsClick={this.showMoreResults}
                    onExpandAllResultsToggle={this.onExpandAllResultsToggle}
                    allExpanded={this.state.allExpanded}
                    showSavedQueryModal={this.state.showSavedQueryModal}
                    onSaveQueryClick={this.showSaveQueryModal}
                    onSavedQueryModalClose={this.onModalClose}
                    onDidCreateSavedQuery={this.onDidCreateSavedQuery}
                    didSave={this.state.didSaveQuery}
                />
            </div>
        )
    }

    /** Combines dynamic filters and search scopes into a list de-duplicated by value. */
    private getFilters(): SearchScopeWithOptionalName[] {
        const filters = new Map<string, SearchScopeWithOptionalName>()

        if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
            let dynamicFilters = this.state.resultsOrError.dynamicFilters
            dynamicFilters = this.state.resultsOrError.dynamicFilters.filter(filter => filter.kind !== 'repo')
            for (const d of dynamicFilters) {
                filters.set(d.value, d)
            }
        }
        const scopes =
            (isSettingsValid<Settings>(this.props.settingsCascade) &&
                this.props.settingsCascade.final['search.scopes']) ||
            []
        if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
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
        // eslint-disable-next-line @typescript-eslint/no-base-to-string
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

        submitSearch(this.props.history, newQuery, 'filter', this.props.patternType, this.props.caseSensitive)
    }
}
