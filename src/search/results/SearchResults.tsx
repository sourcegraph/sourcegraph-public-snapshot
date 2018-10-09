import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import { parseSearchURLQuery, SearchOptions } from '..'
import * as GQL from '../../backend/graphqlschema'
import { PageTitle } from '../../components/PageTitle'
import { currentConfiguration } from '../../settings/configuration'
import { eventLogger } from '../../tracking/eventLogger'
import { search } from '../backend'
import { FilterChip } from '../FilterChip'
import { isSearchResults, submitSearch, toggleSearchFilter } from '../helpers'
import { queryTelemetryData } from '../queryTelemetry'
import { SearchResultsList } from './SearchResultsList'
import { SearchResultsListOld } from './SearchResultsListOld'

const UI_PAGE_SIZE = 75

interface SearchResultsProps {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    navbarSearchQuery: string
    onHelpPopoverToggle: () => void
}

interface SearchScope {
    name?: string
    value: string
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
    /** All search scopes from configuration */
    scopes?: SearchScope[]
}

const newRepoFilters = localStorage.getItem('newRepoFilters') !== 'false'
const newSearchResultsList = localStorage.getItem('newSearchResultsList') !== 'false'

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
        eventLogger.logViewEvent('SearchResults')

        this.subscriptions.add(
            currentConfiguration
                .pipe(map(config => config['search.scopes'] || []))
                .subscribe(scopes => this.setState({ scopes }))
        )

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => parseSearchURLQuery(props.location.search)),
                    // Search when a new search query was specified in the URL
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    filter((searchOptions): searchOptions is SearchOptions => !!searchOptions),
                    tap(searchOptions => {
                        eventLogger.log('SearchResultsQueried', {
                            code_search: { query_data: queryTelemetryData(searchOptions) },
                        })
                    }),
                    switchMap(searchOptions =>
                        concat(
                            // Reset view state
                            [{ resultsOrError: undefined, didSave: false }],
                            // Do async search request
                            search(searchOptions).pipe(
                                // Log telemetry
                                tap(
                                    results =>
                                        eventLogger.log('SearchResultsFetched', {
                                            code_search: {
                                                // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                                results: {
                                                    results_count: results.results.length,
                                                    any_cloning: results.cloning.length > 0,
                                                },
                                            },
                                        }),
                                    error => {
                                        eventLogger.log('SearchResultsFetchFailed', {
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
        eventLogger.log('SavedQueryCreated')
        this.setState({ showSavedQueryModal: false, didSaveQuery: true })
    }

    private onModalClose = () => {
        eventLogger.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        this.setState({ didSaveQuery: false, showSavedQueryModal: false })
    }

    public render(): JSX.Element | null {
        const searchOptions = parseSearchURLQuery(this.props.location.search)
        const filters = this.getFilters()
        return (
            <div className="search-results">
                <PageTitle key="page-title" title={searchOptions && searchOptions.query} />
                {isSearchResults(this.state.resultsOrError) &&
                    filters.length > 0 && (
                        <div className="search-results__filters-bar">
                            Filters:
                            <div className="search-results__filters">
                                {filters
                                    .filter(filter => filter.value !== '')
                                    .map((filter, i) => (
                                        <FilterChip
                                            query={this.props.navbarSearchQuery}
                                            onFilterChosen={this.onDynamicFilterClicked}
                                            key={filter.value}
                                            value={filter.value}
                                            name={filter.name}
                                        />
                                    ))}
                            </div>
                        </div>
                    )}
                {newRepoFilters &&
                    isSearchResults(this.state.resultsOrError) &&
                    this.state.resultsOrError.dynamicFilters.filter(filter => filter.kind === 'repo').length > 0 && (
                        <div className="search-results__filters-bar">
                            Repositories:
                            <div className="search-results__filters">
                                {this.state.resultsOrError.dynamicFilters
                                    .filter(filter => filter.kind === 'repo' && filter.value !== '')
                                    .map((filter, i) => (
                                        <FilterChip
                                            name={filter.label}
                                            query={this.props.navbarSearchQuery}
                                            onFilterChosen={this.onDynamicFilterClicked}
                                            key={filter.value}
                                            value={filter.value}
                                            count={filter.count}
                                            limitHit={filter.limitHit}
                                        />
                                    ))}
                                {this.state.resultsOrError.limitHit &&
                                    !/\brepo:/.test(this.props.navbarSearchQuery) && (
                                        <FilterChip
                                            name="Show more"
                                            query={this.props.navbarSearchQuery}
                                            onFilterChosen={this.showMoreResults}
                                            key={`count:${this.calculateCount()}`}
                                            value={`count:${this.calculateCount()}`}
                                            showMore={true}
                                        />
                                    )}
                            </div>
                        </div>
                    )}
                {newSearchResultsList ? (
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
                        user={this.props.user}
                        isLightTheme={this.props.isLightTheme}
                        onHelpPopoverToggle={this.props.onHelpPopoverToggle}
                    />
                ) : (
                    <SearchResultsListOld
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
                        user={this.props.user}
                        isLightTheme={this.props.isLightTheme}
                        onHelpPopoverToggle={this.props.onHelpPopoverToggle}
                        uiLimit={this.state.uiLimit}
                    />
                )}
            </div>
        )
    }

    /** Combines dynamic filters and search scopes into a list by value. */
    private getFilters(): SearchScope[] {
        const filters = new Map<string, SearchScope>()

        if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
            let dynamicFilters = this.state.resultsOrError.dynamicFilters
            if (newRepoFilters) {
                dynamicFilters = this.state.resultsOrError.dynamicFilters.filter(filter => filter.kind !== 'repo')
            }
            for (const d of dynamicFilters) {
                filters.set(d.value, d)
            }
        }
        if (this.state.scopes) {
            if (isSearchResults(this.state.resultsOrError) && this.state.resultsOrError.dynamicFilters) {
                for (const scope of this.state.scopes) {
                    if (!filters.has(scope.value)) {
                        filters.set(scope.value, scope)
                    }
                }
            } else {
                for (const scope of this.state.scopes) {
                    // Check for duplicated filters and prioritize the user's configured scope names
                    const duplicatedFilter = filters.get(scope.value)

                    if (duplicatedFilter && scope.name) {
                        duplicatedFilter.name = scope.name
                    } else if (duplicatedFilter && !scope.name) {
                        // Show scope value if user doesn't provide a scope name
                        duplicatedFilter.name = scope.value
                    }
                    filters.set(scope.value, duplicatedFilter ? duplicatedFilter : scope)
                }
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
                eventLogger.log(this.state.allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
            }
        )
    }

    private onDynamicFilterClicked = (value: string) => {
        eventLogger.log('DynamicFilterClicked', {
            search_filter: { value },
        })
        submitSearch(this.props.history, { query: toggleSearchFilter(this.props.navbarSearchQuery, value) }, 'filter')
    }
}
