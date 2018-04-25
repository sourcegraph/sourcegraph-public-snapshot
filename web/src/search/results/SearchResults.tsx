import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, startWith, switchMap, tap } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { eventLogger } from '../../tracking/eventLogger'
import { search } from './../backend'
import { FilterChip } from './../FilterChip'
import { isSearchResults } from './../helpers'
import { parseSearchURLQuery, SearchOptions } from './../index'
import { queryTelemetryData } from './../queryTelemetry'
import { SearchResultsList } from './SearchResultsList'

const ALL_EXPANDED_LOCAL_STORAGE_KEY = 'allExpanded'
const UI_PAGE_SIZE = 75

interface SearchResultsProps {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    navbarSearchQuery: string
    onFilterChosen: (value: string) => void
}

interface SearchResultsState {
    /** The loaded search results, error or undefined while loading */
    resultsOrError?: GQL.ISearchResults
    allExpanded: boolean
    uiLimit: number

    // Saved Queries
    showSavedQueryModal: boolean
    didSaveQuery: boolean
}

export class SearchResults extends React.Component<SearchResultsProps, SearchResultsState> {
    public state: SearchResultsState = {
        didSaveQuery: false,
        showSavedQueryModal: false,
        allExpanded: localStorage.getItem(ALL_EXPANDED_LOCAL_STORAGE_KEY) === 'true',
        uiLimit: UI_PAGE_SIZE,
    }

    /** Emits on componentDidUpdate with the new props */
    private componentUpdates = new Subject<SearchResultsProps>()

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SearchResults')

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
                            [{ resultsOrError: undefined, didSave: false, uiLimit: UI_PAGE_SIZE }],
                            // Do async search request
                            search(searchOptions).pipe(
                                // Log telemetry
                                tap(
                                    results =>
                                        eventLogger.log('SearchResultsFetched', {
                                            code_search: {
                                                // 🚨 PRIVACY: never provide any private data in { code_search: { results } }.
                                                // This field is whitelisted for on-premises Server users.
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
                                            key={filter.value}
                                            value={filter.value}
                                        />
                                    ))}
                            </div>
                        </div>
                    )}
                <SearchResultsList
                    resultsOrError={this.state.resultsOrError}
                    uiLimit={this.state.uiLimit}
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

    private onExpandAllResultsToggle = () => {
        this.setState(
            state => ({ allExpanded: !state.allExpanded }),
            () => {
                localStorage.setItem(ALL_EXPANDED_LOCAL_STORAGE_KEY, this.state.allExpanded + '')
                eventLogger.log(this.state.allExpanded ? 'allResultsExpanded' : 'allResultsCollapsed')
            }
        )
    }

    private onDynamicFilterClicked = (value: string) => {
        eventLogger.log('DynamicFilterClicked', {
            search_filter: { value },
        })
        this.props.onFilterChosen(value)
    }
}
