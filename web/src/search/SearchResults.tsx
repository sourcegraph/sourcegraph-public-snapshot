import * as H from 'history'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { eventLogger } from '../tracking/eventLogger'
import { search } from './backend'
import { FilterChip } from './FilterChip'
import { isSearchResults } from './helpers'
import { parseSearchURLQuery, SearchOptions, searchOptionsEqual } from './index'
import { queryTelemetryData } from './queryTelemetry'
import { SearchResultsList } from './SearchResultsList'

const ALL_EXPANDED_LOCAL_STORAGE_KEY = 'allExpanded'
const UI_PAGE_SIZE = 75

interface SearchResultsProps {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onFilterChosen: (value: string) => void
    navbarSearchQuery: string
}

interface SearchResultsState {
    resultsOrError?: GQL.ISearchResults
    showModal: boolean
    didSave: boolean
    allExpanded: boolean
    uiLimit: number
}

export class SearchResults extends React.Component<SearchResultsProps, SearchResultsState> {
    public state: SearchResultsState = {
        didSave: false,
        showModal: false,
        allExpanded: localStorage.getItem(ALL_EXPANDED_LOCAL_STORAGE_KEY) === 'true',
        uiLimit: UI_PAGE_SIZE,
    }

    private componentUpdates = new Subject<SearchResultsProps>()
    private searchRequested = new Subject<SearchOptions>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SearchResults')

        this.subscriptions.add(
            this.searchRequested
                // Don't search using stale search options.
                .pipe(
                    filter(searchOptions => {
                        const currentSearchOptions = parseSearchURLQuery(this.props.location.search)
                        return !currentSearchOptions || searchOptionsEqual(searchOptions, currentSearchOptions)
                    }),
                    switchMap(searchOptions => {
                        eventLogger.log('SearchResultsQueried', {
                            code_search: { query_data: queryTelemetryData(searchOptions) },
                        })

                        return search(searchOptions).pipe(
                            tap(
                                res =>
                                    eventLogger.log('SearchResultsFetched', {
                                        code_search: {
                                            // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                            // This field is whitelisted for on-premises Server users.
                                            results: {
                                                results_count: res.results.length,
                                                result_items_count: res.results.reduce(
                                                    (count, result) => count + resultItemsCount(result),
                                                    0
                                                ),
                                                any_cloning: res.cloning.length > 0,
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
                            map(results => ({ resultsOrError: results, uiLimit: UI_PAGE_SIZE })),
                            catchError(error => [
                                {
                                    resultsOrError: error,
                                    didSave: false,
                                    showModal: false,
                                    allExpanded: false,
                                    uiLimit: UI_PAGE_SIZE,
                                },
                            ])
                        )
                    })
                )
                .subscribe(newState => this.setState(newState as SearchResultsState), err => console.error(err))
        )
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => props.location),
                    distinctUntilChanged(),
                    tap(location => {
                        const searchOptions = parseSearchURLQuery(location.search)
                        setTimeout(() => this.searchRequested.next(searchOptions))
                    }),
                    map(() => ({
                        resultsOrError: undefined,
                        didSave: false,
                        showModal: false,
                        allExpanded: localStorage.getItem(ALL_EXPANDED_LOCAL_STORAGE_KEY) === 'true',
                        uiLimit: UI_PAGE_SIZE,
                    }))
                )
                .subscribe(newState => this.setState(newState as SearchResultsState), err => console.error(err))
        )
    }

    public componentWillReceiveProps(newProps: SearchResultsProps): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private showSaveQueryModal = () => {
        this.setState({ showModal: true, didSave: false })
    }

    private onDidCreateSavedQuery = () => {
        eventLogger.log('SavedQueryCreated')
        this.setState({ showModal: false, didSave: true })
    }

    private onModalClose = () => {
        eventLogger.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        this.setState({ didSave: false, showModal: false })
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-results">
                {isSearchResults(this.state.resultsOrError) &&
                    this.state.resultsOrError.dynamicFilters.length > 0 && (
                        <div className="search-results__filters-bar">
                            Filters:
                            <div className="search-results__filters">
                                {this.state.resultsOrError.dynamicFilters
                                    .filter(filter => filter.value !== '')
                                    .map((filter, i) => (
                                        <FilterChip
                                            query={this.props.navbarSearchQuery}
                                            onFilterChosen={this.onDynamicFilterClicked}
                                            key={i}
                                            value={filter.value}
                                        />
                                    ))}
                            </div>
                        </div>
                    )}
                <SearchResultsList
                    resultsOrError={this.state.resultsOrError}
                    showModal={this.state.showModal}
                    onDidCreateSavedQuery={this.onDidCreateSavedQuery}
                    onModalClose={this.onModalClose}
                    onExpandAllResultsClick={this.expandAllResults}
                    onShowMoreResultsClick={this.showMoreResults}
                    allExpanded={this.state.allExpanded}
                    isLightTheme={this.props.isLightTheme}
                    location={this.props.location}
                    user={this.props.user}
                    onSaveQueryClick={this.showSaveQueryModal}
                    didSave={this.state.didSave}
                    uiLimit={this.state.uiLimit}
                />
            </div>
        )
    }

    private showMoreResults = () => {
        // This function can only get called if the results were successfully loaded,
        // so casting is the right thing to do here
        const results = this.state.resultsOrError as GQL.ISearchResults
        if (results.results.length > this.state.uiLimit) {
            // We already have results fetched that aren't being displayed.
            // Increase the UI limit and rerender.
            this.setState(state => ({ uiLimit: state.uiLimit + UI_PAGE_SIZE }))
            return
        }

        // Requery with an increased max result count.
        const params = new URLSearchParams(this.props.location.search)
        let query = params.get('q') || ''

        const defaultMaxSearchResults = Math.max(results.resultCount || 0, 30)

        const m = query.match(/max:(\d+)/)
        if (m) {
            let n = parseInt(m[1], 10)
            if (!(n >= 1)) {
                n = defaultMaxSearchResults
            }
            query = query.replace(/max:\d+/g, '').trim() + ` max:${n * 2}`
        } else {
            query = `${query} max:${defaultMaxSearchResults}`
        }
        params.set('q', query)
        this.props.history.replace({ search: params.toString() })
    }

    private expandAllResults = () => {
        const allExpanded = !this.state.allExpanded
        localStorage.setItem(ALL_EXPANDED_LOCAL_STORAGE_KEY, allExpanded + '')
        this.setState(
            state => ({ allExpanded }),
            () => {
                eventLogger.log(allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
            }
        )
    }

    private onDynamicFilterClicked = (value: string) => {
        eventLogger.log('DynamicFilterClicked', {
            search_filter: {
                value,
            },
        })
        this.props.onFilterChosen(value)
    }
}

function resultItemsCount(result: GQL.SearchResult): number {
    switch (result.__typename) {
        case 'FileMatch':
            return 1
        case 'CommitSearchResult':
            return 1
    }
    return 1
}
