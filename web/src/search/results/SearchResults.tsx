import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import { parseSearchURLQuery } from '..'
import { EvaluatedContributions } from '../../../../shared/src/api/protocol'
import { FetchFileCtx } from '../../../../shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { PageTitle } from '../../components/PageTitle'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../theme'
import { EventLogger } from '../../tracking/eventLogger'
import {
    isSearchResults,
    submitSearch,
    toggleSearchFilter,
    toggleSearchFilterAndReplaceSampleRepogroup,
} from '../helpers'
import { queryTelemetryData } from '../queryTelemetry'
import { SearchResultsFilterBars, SearchScopeWithOptionalName } from './SearchResultsFilterBars'
import { SearchResultsList } from './SearchResultsList'

const UI_PAGE_SIZE = 75

export interface SearchResultsProps extends ExtensionsControllerProps<'services'>, SettingsCascadeProps, ThemeProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    navbarSearchQuery: string
    telemetryService: Pick<EventLogger, 'log' | 'logViewEvent'>
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
    searchRequest: (
        query: string,
        { extensionsController }: ExtensionsControllerProps<'services'>
    ) => Observable<GQL.ISearchResults | ErrorLike>
    isSourcegraphDotCom: boolean
    deployType: DeployType
}

interface SearchResultsState {
    /** The loaded search results, error or undefined while loading */
    resultsOrError?: GQL.ISearchResults
    allExpanded: boolean

    // TODO: Remove when newSearchResultsList is removed
    uiLimit: number

    // Saved Queries
    showSavedQueryModal: boolean
    didSaveQuery: boolean
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: EvaluatedContributions
}

export class SearchResults extends React.Component<SearchResultsProps, SearchResultsState> {
    public state: SearchResultsState = {
        didSaveQuery: false,
        showSavedQueryModal: false,
        allExpanded: false,
        uiLimit: UI_PAGE_SIZE,
    }
    /** Emits on componentDidUpdate with the new props */
    private componentUpdates = new Subject<SearchResultsProps>()

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.props.telemetryService.logViewEvent('SearchResults')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => parseSearchURLQuery(props.location.search)),
                    // Search when a new search query was specified in the URL
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    filter((query): query is string => !!query),
                    tap(query => {
                        const query_data = queryTelemetryData(query)
                        this.props.telemetryService.log('SearchResultsQueried', {
                            code_search: { query_data },
                        })
                        if (
                            query_data.query &&
                            query_data.query.field_type &&
                            query_data.query.field_type.value_diff > 0
                        ) {
                            this.props.telemetryService.log('DiffSearchResultsQueried')
                        }
                    }),
                    switchMap(query =>
                        concat(
                            // Reset view state
                            [{ resultsOrError: undefined, didSave: false }],
                            // Do async search request
                            this.props.searchRequest(query, this.props).pipe(
                                // Log telemetry
                                tap(
                                    results =>
                                        this.props.telemetryService.log('SearchResultsFetched', {
                                            code_search: {
                                                // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                                results: {
                                                    results_count: isErrorLike(results) ? 0 : results.results.length,
                                                    any_cloning: isErrorLike(results)
                                                        ? false
                                                        : results.cloning.length > 0,
                                                },
                                            },
                                        }),
                                    error => {
                                        this.props.telemetryService.log('SearchResultsFetchFailed', {
                                            code_search: { error_message: error.message },
                                        })
                                        console.error(error)
                                    }
                                ),
                                // Update view with results or error
                                map(results => ({ resultsOrError: results })),
                                catchError(error => [{ resultsOrError: error }])
                            )
                        )
                    )
                )
                .subscribe(newState => this.setState(newState as SearchResultsState), err => console.error(err))
        )

        this.props.extensionsController.services.contribution
            .getContributions()
            .subscribe(contributions => this.setState({ contributions }))
    }

    public componentDidUpdate(prevProps: SearchResultsProps): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private showSaveQueryModal = () => {
        this.setState({ showSavedQueryModal: true, didSaveQuery: false })
    }

    private onDidCreateSavedQuery = () => {
        this.props.telemetryService.log('SavedQueryCreated')
        this.setState({ showSavedQueryModal: false, didSaveQuery: true })
    }

    private onModalClose = () => {
        this.props.telemetryService.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        this.setState({ didSaveQuery: false, showSavedQueryModal: false })
    }

    public render(): JSX.Element | null {
        const query = parseSearchURLQuery(this.props.location.search)
        const filters = this.getFilters()
        const extensionFilters = this.state.contributions && this.state.contributions.searchFilters

        return (
            <div className="search-results d-flex flex-column w-100">
                <PageTitle key="page-title" title={query} />
                <SearchResultsFilterBars
                    navbarSearchQuery={this.props.navbarSearchQuery}
                    results={this.state.resultsOrError}
                    filters={filters}
                    extensionFilters={extensionFilters}
                    onFilterClick={this.onDynamicFilterClicked}
                    onShowMoreResultsClick={this.showMoreResults}
                    calculateShowMoreResultsCount={this.calculateCount}
                />
                <SearchResultsList
                    resultsOrError={this.state.resultsOrError}
                    onShowMoreResultsClick={this.showMoreResults}
                    onExpandAllResultsToggle={this.onExpandAllResultsToggle}
                    allExpanded={this.state.allExpanded}
                    showSavedQueryModal={this.state.showSavedQueryModal}
                    onSaveQueryClick={this.showSaveQueryModal}
                    onSavedQueryModalClose={this.onModalClose}
                    onDidCreateSavedQuery={this.onDidCreateSavedQuery}
                    didSave={this.state.didSaveQuery}
                    location={this.props.location}
                    history={this.props.history}
                    authenticatedUser={this.props.authenticatedUser}
                    settingsCascade={this.props.settingsCascade}
                    isLightTheme={this.props.isLightTheme}
                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                    fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
                    deployType={this.props.deployType}
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
    private showMoreResults = () => {
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
            return Math.max(results.resultCount * 2, 1000)
        }
        return Math.max(results.resultCount * 2 || 0, 1000)
    }

    private onExpandAllResultsToggle = () => {
        this.setState(
            state => ({ allExpanded: !state.allExpanded }),
            () => {
                this.props.telemetryService.log(this.state.allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
            }
        )
    }

    private onDynamicFilterClicked = (value: string) => {
        this.props.telemetryService.log('DynamicFilterClicked', {
            search_filter: { value },
        })

        const newQuery = this.props.isSourcegraphDotCom
            ? toggleSearchFilterAndReplaceSampleRepogroup(this.props.navbarSearchQuery, value)
            : toggleSearchFilter(this.props.navbarSearchQuery, value)

        submitSearch(this.props.history, newQuery, 'filter')
    }
}
