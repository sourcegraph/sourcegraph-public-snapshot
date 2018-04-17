import ArrowCollapseVerticalIcon from '@sourcegraph/icons/lib/ArrowCollapseVertical'
import ArrowExpandVerticalIcon from '@sourcegraph/icons/lib/ArrowExpandVertical'
import CalculatorIcon from '@sourcegraph/icons/lib/Calculator'
import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import DirectionalSign from '@sourcegraph/icons/lib/DirectionalSign'
import DocumentIcon from '@sourcegraph/icons/lib/Document'
import DownloadIcon from '@sourcegraph/icons/lib/Download'
import HourglassIcon from '@sourcegraph/icons/lib/Hourglass'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import SaveIcon from '@sourcegraph/icons/lib/Save'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
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
import { ServerBanner } from '../marketing/ServerBanner'
import { eventLogger } from '../tracking/eventLogger'
import { ErrorLike, isErrorLike } from '../util/errors'
import { showDotComMarketing } from '../util/features'
import { pluralize } from '../util/strings'
import { search } from './backend'
import { CommitSearchResult } from './CommitSearchResult'
import { FileMatch } from './FileMatch'
import { FilterChip } from './FilterChip'
import { parseSearchURLQuery, SearchOptions, searchOptionsEqual } from './index'
import { ModalContainer } from './ModalContainer'
import { queryTelemetryData } from './queryTelemetry'
import { RepositorySearchResult } from './RepositorySearchResult'
import { SavedQueryCreateForm } from './SavedQueryCreateForm'
import { SearchAlert } from './SearchAlert'

const ALL_EXPANDED_LOCAL_STORAGE_KEY = 'allExpanded'
const DATA_CENTER_UPGRADE_STRING =
    'Upgrade to Sourcegraph Data Center for distributed on-the-fly search and near-instant indexed search.'
const SEARCH_TIMED_OUT_DEFAULT_TITLE = 'Search timed out'
const UI_PAGE_SIZE = 75

interface SearchResultsListProps {
    isLightTheme: boolean
    location: H.Location
    user: GQL.IUser | null

    resultsOrError?: GQL.ISearchResults | ErrorLike
    uiLimit: number
    onShowMoreResultsClick: () => void

    allExpanded?: boolean
    onExpandAllResultsClick: () => void

    // Saved queries
    showModal: boolean
    onModalClose: () => void
    onDidCreateSavedQuery: () => void
    onSaveQueryClick: () => void
    didSave?: boolean
}

const isSearchResults = (val: any): val is GQL.ISearchResults =>
    val && typeof val === 'object' && val.__typename === 'SearchResults'

class SearchResultsList extends React.PureComponent<SearchResultsListProps, {}> {
    public render(): React.ReactNode {
        let alert: {
            title: string
            description?: string | null
            proposedQueries?: GQL.ISearchQueryDescription[]
            errorBody?: React.ReactFragment
        } | null = null
        const searchTimeoutParameterEnabled = window.context.searchTimeoutParameterEnabled
        if (this.props.resultsOrError) {
            if (isErrorLike(this.props.resultsOrError)) {
                const error = this.props.resultsOrError
                if (error.message.includes('no query terms or regexp specified')) {
                    alert = { title: '', description: 'Enter terms to search...' }
                } else {
                    alert = { title: 'Something went wrong', description: upperFirst(error.message) }
                }
            } else {
                const results = this.props.resultsOrError
                if (results.alert) {
                    alert = results.alert
                } else if (
                    results.results.length === 0 &&
                    results.missing.length === 0 &&
                    results.cloning.length === 0
                ) {
                    const defaultTimeoutAlert = {
                        title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                        description: searchTimeoutParameterEnabled
                            ? "Try narrowing your query, or specifying a longer 'timeout:' in your query."
                            : 'Try narrowing your query.',
                    }
                    const longerTimeoutString = searchTimeoutParameterEnabled
                        ? "Specify a longer 'timeout:' in your query."
                        : ''
                    if (results.timedout.length > 0) {
                        if (window.context.sourcegraphDotComMode) {
                            alert = defaultTimeoutAlert
                        } else {
                            if (window.context.likelyDockerOnMac) {
                                alert = {
                                    title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                                    errorBody: this.renderSearchAlertTimeoutDetails([
                                        longerTimeoutString,
                                        DATA_CENTER_UPGRADE_STRING,
                                        'Use Docker Machine instead of Docker for Mac for better performance on macOS.',
                                    ]),
                                }
                            } else if (!window.context.likelyDockerOnMac && !window.context.isRunningDataCenter) {
                                alert = {
                                    title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                                    errorBody: this.renderSearchAlertTimeoutDetails([
                                        longerTimeoutString,
                                        DATA_CENTER_UPGRADE_STRING,
                                        'Run Sourcegraph on a server with more CPU and memory, or faster disk IO.',
                                    ]),
                                }
                            } else {
                                alert = defaultTimeoutAlert
                            }
                        }
                    } else {
                        alert = { title: 'No results' }
                    }
                }
            }
        }

        const parsedQuery = parseSearchURLQuery(this.props.location.search)
        const showMissingReposEnabled =
            window.context.showMissingReposEnabled || localStorage.getItem('showMissingRepos')

        return (
            <div className="search-results__list">
                {/* Saved Queries Form */}
                {this.props.showModal && (
                    <ModalContainer
                        onClose={this.props.onModalClose}
                        component={
                            <SavedQueryCreateForm
                                user={this.props.user}
                                values={{ query: parsedQuery ? parsedQuery.query : '' }}
                                onDidCancel={this.props.onModalClose}
                                onDidCreate={this.props.onDidCreateSavedQuery}
                            />
                        }
                    />
                )}

                {/* Loader */}
                {this.props.resultsOrError === undefined && <Loader className="icon-inline" />}

                {isSearchResults(this.props.resultsOrError) &&
                    (() => {
                        const results = this.props.resultsOrError
                        return (
                            <>
                                {/* Info Bar */}
                                <div className="search-results__info">
                                    {(results.timedout.length > 0 ||
                                        results.cloning.length > 0 ||
                                        results.results.length > 0 ||
                                        (showMissingReposEnabled && results.missing.length > 0)) && (
                                        <small className="search-results__info-row">
                                            <div className="search-results__info-row-left">
                                                {/* Time stats */}
                                                {
                                                    <div className="search-results__notice e2e-search-results-stats">
                                                        <span>
                                                            <CalculatorIcon className="icon-inline" />{' '}
                                                            {results.approximateResultCount}{' '}
                                                            {pluralize('result', results.resultCount)} in{' '}
                                                            {(results.elapsedMilliseconds / 1000).toFixed(2)} seconds
                                                            {results.indexUnavailable && ' (index unavailable)'}
                                                        </span>
                                                    </div>
                                                }
                                                {/* Missing repos */}
                                                {showMissingReposEnabled &&
                                                    results.missing.length > 0 && (
                                                        <div
                                                            className="search-results__notice"
                                                            data-tooltip={results.missing.join('\n')}
                                                        >
                                                            <span>
                                                                <DirectionalSign className="icon-inline" />{' '}
                                                                {results.missing.length}{' '}
                                                                {pluralize(
                                                                    'repository',
                                                                    results.missing.length,
                                                                    'repositories'
                                                                )}{' '}
                                                                not found
                                                            </span>
                                                        </div>
                                                    )}
                                                {/* Timed out repos */}
                                                {results.timedout.length > 0 && (
                                                    <div
                                                        className="search-results__notice"
                                                        data-tooltip={results.timedout.join('\n')}
                                                    >
                                                        <span>
                                                            <HourglassIcon className="icon-inline" />{' '}
                                                            {results.timedout.length}{' '}
                                                            {pluralize(
                                                                'repository',
                                                                results.timedout.length,
                                                                'repositories'
                                                            )}{' '}
                                                            timed out (reload to try again, or specify a longer
                                                            "timeout:" in your query)
                                                        </span>
                                                    </div>
                                                )}
                                                {/* Cloning repos */}
                                                {results.cloning.length > 0 && (
                                                    <div
                                                        className="search-results__notice"
                                                        data-tooltip={results.cloning.join('\n')}
                                                    >
                                                        <span>
                                                            <DownloadIcon className="icon-inline" />{' '}
                                                            {results.cloning.length}{' '}
                                                            {pluralize(
                                                                'repository',
                                                                results.cloning.length,
                                                                'repositories'
                                                            )}{' '}
                                                            cloning (reload to try again)
                                                        </span>
                                                    </div>
                                                )}
                                            </div>
                                            <div className="search-results__info-row-right">
                                                <button
                                                    onClick={this.props.onExpandAllResultsClick}
                                                    className="btn btn-link"
                                                >
                                                    {this.props.allExpanded ? (
                                                        <>
                                                            <ArrowCollapseVerticalIcon
                                                                className="icon-inline"
                                                                data-tooltip="Collapse"
                                                            />{' '}
                                                            Collapse all
                                                        </>
                                                    ) : (
                                                        <>
                                                            <ArrowExpandVerticalIcon
                                                                className="icon-inline"
                                                                data-tooltip="Expand"
                                                            />{' '}
                                                            Expand all
                                                        </>
                                                    )}
                                                </button>
                                                {!this.props.didSave &&
                                                    this.props.user && (
                                                        <button
                                                            onClick={this.props.onSaveQueryClick}
                                                            className="btn btn-link"
                                                        >
                                                            <SaveIcon className="icon-inline" /> Save this search query
                                                        </button>
                                                    )}
                                                {this.props.didSave && (
                                                    <span>
                                                        <CheckmarkIcon className="icon-inline" /> Query saved
                                                    </span>
                                                )}
                                            </div>
                                        </small>
                                    )}
                                    {!results.alert && showDotComMarketing && <ServerBanner />}
                                </div>

                                {/* Results */}
                                {results.results
                                    .slice(0, this.props.uiLimit)
                                    .map((result, i) => this.renderResult(i, result, i <= 15))}

                                {/* Show more button */}
                                {(results.limitHit || results.results.length > this.props.uiLimit) && (
                                    <button
                                        className="btn btn-link search-results__more"
                                        onClick={this.props.onShowMoreResultsClick}
                                    >
                                        Show more
                                    </button>
                                )}
                            </>
                        )
                    })()}
                {alert && (
                    <SearchAlert
                        className="search-results__alert"
                        title={alert.title}
                        description={alert.description || undefined}
                        proposedQueries={alert.proposedQueries}
                        location={this.props.location}
                        errorBody={alert.errorBody}
                    />
                )}
            </div>
        )
    }

    private renderSearchAlertTimeoutDetails(items: string[]): React.ReactFragment {
        return (
            <div className="search-alert__list">
                <p className="search-alert__list-header">Recommendations:</p>
                <ul className="search-alert__list-items">
                    {items.map(
                        (item, i) =>
                            item && (
                                <li key={i} className="search-alert__list-item">
                                    {item}
                                </li>
                            )
                    )}
                </ul>
            </div>
        )
    }

    private renderResult(key: number, result: GQL.SearchResult, expanded: boolean): React.ReactNode {
        switch (result.__typename) {
            case 'Repository':
                return <RepositorySearchResult key={key} result={result} onSelect={this.logEvent} />
            case 'FileMatch':
                return (
                    <FileMatch
                        key={key}
                        icon={result.lineMatches && result.lineMatches.length > 0 ? RepoIcon : DocumentIcon}
                        result={result}
                        onSelect={this.logEvent}
                        expanded={false}
                        showAllMatches={false}
                        isLightTheme={this.props.isLightTheme}
                        allExpanded={this.props.allExpanded}
                    />
                )
            case 'CommitSearchResult':
                return (
                    <CommitSearchResult
                        key={key}
                        location={this.props.location}
                        result={result}
                        onSelect={this.logEvent}
                        expanded={expanded}
                        allExpanded={this.props.allExpanded}
                    />
                )
        }
        return undefined
    }

    private logEvent = () => eventLogger.log('SearchResultClicked')
}

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
    didSave?: boolean
    allExpanded?: boolean
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
