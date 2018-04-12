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
import { currentUser } from '../auth'
import { ServerBanner } from '../marketing/ServerBanner'
import { eventLogger } from '../tracking/eventLogger'
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

    // Saved queries
    showModal: boolean
    onModalClose: () => void
    onDidCreateSavedQuery: () => void
    onSaveQueryClick: () => void

    results: GQL.IFileMatch[]
    cloning: string[]
    missing: string[]
    timedout: string[]
    resultCount?: number
    approximateResultCount?: string
    alert: GQL.ISearchAlert | null
    elapsedMilliseconds?: number
    loading: boolean
    error?: Error
    limitHit: boolean
    indexUnavailable: boolean
    didSave?: boolean
    allExpanded?: boolean
    uiLimit: number
    onExpandAllResultsClick: () => void
    onShowMoreResultsClick: () => void
}

class SearchResultsList extends React.PureComponent<SearchResultsListProps, {}> {
    public render(): React.ReactNode {
        let alert: {
            title: string
            description?: string | null
            proposedQueries?: GQL.ISearchQueryDescription[]
            errorBody?: React.ReactFragment
        } | null = null
        const searchTimeoutParameterEnabled = window.context.searchTimeoutParameterEnabled
        if (this.props.error) {
            if (this.props.error.message.includes('no query terms or regexp specified')) {
                alert = { title: '', description: 'Enter terms to search...' }
            } else {
                alert = { title: 'Something went wrong', description: upperFirst(this.props.error.message) }
            }
        } else if (this.props.alert) {
            alert = this.props.alert
        } else if (
            !this.props.loading &&
            this.props.results.length === 0 &&
            this.props.missing.length === 0 &&
            this.props.cloning.length === 0
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
            if (this.props.timedout.length > 0) {
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

        const parsedQuery = parseSearchURLQuery(this.props.location.search)
        const showMissingReposEnabled =
            window.context.showMissingReposEnabled || localStorage.getItem('showMissingRepos')
        const showMissingRepos = showMissingReposEnabled && this.props.missing.length > 0

        return (
            <div className="search-results__list">
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
                <div className="search-results__info">
                    {(this.props.timedout.length > 0 ||
                        this.props.cloning.length > 0 ||
                        this.props.results.length > 0 ||
                        showMissingRepos) && (
                        <small className="search-results__info-row">
                            <div className="search-results__info-row-left">
                                {/* Time stats */}
                                {typeof this.props.approximateResultCount === 'string' &&
                                    typeof this.props.resultCount === 'number' && (
                                        <div className="search-results__notice e2e-search-results-stats">
                                            <span>
                                                <CalculatorIcon className="icon-inline" />{' '}
                                                {this.props.approximateResultCount}{' '}
                                                {pluralize('result', this.props.resultCount)}
                                                {typeof this.props.elapsedMilliseconds === 'number' && (
                                                    <>
                                                        {' '}
                                                        in {(this.props.elapsedMilliseconds / 1000).toFixed(2)} seconds
                                                    </>
                                                )}
                                                {this.props.indexUnavailable && ' (index unavailable)'}
                                            </span>
                                        </div>
                                    )}
                                {/* Missing repos */}
                                {showMissingRepos && (
                                    <div
                                        className="search-results__notice"
                                        data-tooltip={this.props.missing.join('\n')}
                                    >
                                        <span>
                                            <DirectionalSign className="icon-inline" /> {this.props.missing.length}{' '}
                                            {pluralize('repository', this.props.missing.length, 'repositories')} not{' '}
                                            found
                                        </span>
                                    </div>
                                )}
                                {/* Timed out repos */}
                                {this.props.timedout.length > 0 && (
                                    <div
                                        className="search-results__notice"
                                        data-tooltip={this.props.timedout.join('\n')}
                                    >
                                        <span>
                                            <HourglassIcon className="icon-inline" /> {this.props.timedout.length}{' '}
                                            {pluralize('repository', this.props.timedout.length, 'repositories')} timed{' '}
                                            out (reload to try again, or specify a longer "timeout:" in your query)
                                        </span>
                                    </div>
                                )}
                                {/* Cloning repos */}
                                {this.props.cloning.length > 0 && (
                                    <div
                                        className="search-results__notice"
                                        data-tooltip={this.props.cloning.join('\n')}
                                    >
                                        <span>
                                            <DownloadIcon className="icon-inline" /> {this.props.cloning.length}{' '}
                                            {pluralize('repository', this.props.cloning.length, 'repositories')} cloning{' '}
                                            (reload to try again)
                                        </span>
                                    </div>
                                )}
                            </div>
                            <div className="search-results__info-row-right">
                                <button onClick={this.props.onExpandAllResultsClick} className="btn btn-link">
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
                                            <ArrowExpandVerticalIcon className="icon-inline" data-tooltip="Expand" />{' '}
                                            Expand all
                                        </>
                                    )}
                                </button>
                                {!this.props.didSave &&
                                    this.props.user && (
                                        <button onClick={this.props.onSaveQueryClick} className="btn btn-link">
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
                    {!this.props.alert &&
                        !this.props.error &&
                        !this.props.loading &&
                        showDotComMarketing && <ServerBanner />}
                </div>
                {this.props.results.length > 0}
                {this.props.loading && <Loader className="icon-inline" />}
                {this.props.results
                    .slice(0, this.props.uiLimit)
                    .map((result, i) => this.renderResult(i, result, i <= 15))}
                {(this.props.limitHit || this.props.results.length > this.props.uiLimit) && (
                    <button className="btn btn-link search-results__more" onClick={this.props.onShowMoreResultsClick}>
                        Show more
                    </button>
                )}
                {alert && (
                    <SearchAlert
                        className="search-results__alert"
                        title={alert.title}
                        description={alert.description || undefined}
                        proposedQueries={this.props.alert ? this.props.alert.proposedQueries : undefined}
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
    results: GQL.IFileMatch[]
    resultCount?: number
    approximateResultCount?: string
    alert: GQL.ISearchAlert | null
    elapsedMilliseconds?: number
    loading: boolean
    error?: Error
    limitHit: boolean
    cloning: string[]
    missing: string[]
    timedout: string[]
    indexUnavailable: boolean
    showModal: boolean
    didSave?: boolean
    user?: GQL.IUser | null
    dynamicFilters: GQL.ISearchFilter[]
    allExpanded?: boolean
    uiLimit: number
}

export class SearchResults extends React.Component<SearchResultsProps, SearchResultsState> {
    public state: SearchResultsState = {
        results: [],
        alert: null,
        loading: true,
        limitHit: false,
        cloning: [],
        missing: [],
        timedout: [],
        indexUnavailable: false,
        didSave: false,
        showModal: false,
        dynamicFilters: [],
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
                            map(res => ({
                                ...res,
                                error: undefined,
                                loading: false,
                                uiLimit: UI_PAGE_SIZE,
                            })),
                            catchError((error): SearchResultsState[] => [
                                {
                                    results: [],
                                    resultCount: undefined,
                                    approximateResultCount: undefined,
                                    alert: null,
                                    elapsedMilliseconds: undefined,
                                    missing: [],
                                    cloning: [],
                                    timedout: [],
                                    indexUnavailable: false,
                                    limitHit: false,
                                    error,
                                    loading: false,
                                    didSave: false,
                                    showModal: false,
                                    dynamicFilters: [],
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
                        results: [],
                        resultCount: undefined,
                        approximateResultCount: undefined,
                        alert: null,
                        elapsedMilliseconds: undefined,
                        missing: [],
                        cloning: [],
                        timedout: [],
                        indexUnavailable: false,
                        limitHit: false,
                        error: undefined,
                        loading: true,
                        didSave: false,
                        showModal: false,
                        dynamicFilters: [],
                        allExpanded: localStorage.getItem(ALL_EXPANDED_LOCAL_STORAGE_KEY) === 'true',
                        uiLimit: UI_PAGE_SIZE,
                    }))
                )
                .subscribe(newState => this.setState(newState as SearchResultsState), err => console.error(err))
        )
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
    }

    public componentWillReceiveProps(newProps: SearchResultsProps): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private showSaveQueryModal = () => {
        this.setState({
            showModal: true,
            didSave: false,
        })
    }

    private onDidCreateSavedQuery = () => {
        eventLogger.log('SavedQueryCreated')
        this.setState({
            showModal: false,
            didSave: true,
        })
    }

    private onModalClose = () => {
        eventLogger.log('SavedQueriesToggleCreating', { queries: { creating: false } })
        this.setState({
            didSave: false,
            showModal: false,
        })
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-results">
                {this.state.results.length > 0 &&
                    this.state.dynamicFilters.length > 0 && (
                        <div className="search-results__filters-bar">
                            Filters:
                            <div className="search-results__filters">
                                {this.state.dynamicFilters
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
                    showModal={this.state.showModal}
                    onDidCreateSavedQuery={this.onDidCreateSavedQuery}
                    onModalClose={this.onModalClose}
                    results={this.state.results}
                    timedout={this.state.timedout}
                    missing={this.state.missing}
                    alert={this.state.alert}
                    onExpandAllResultsClick={this.expandAllResults}
                    onShowMoreResultsClick={this.showMoreResults}
                    allExpanded={this.state.allExpanded}
                    approximateResultCount={this.state.approximateResultCount}
                    isLightTheme={this.props.isLightTheme}
                    location={this.props.location}
                    user={this.props.user}
                    onSaveQueryClick={this.showSaveQueryModal}
                    cloning={this.state.cloning}
                    resultCount={this.state.resultCount}
                    elapsedMilliseconds={this.state.elapsedMilliseconds}
                    loading={this.state.loading}
                    error={this.state.error}
                    limitHit={this.state.limitHit}
                    indexUnavailable={this.state.indexUnavailable}
                    didSave={this.state.didSave}
                    uiLimit={this.state.uiLimit}
                />
            </div>
        )
    }

    private showMoreResults = () => {
        if (this.state.results.length > this.state.uiLimit) {
            // We already have results fetched that aren't being displayed.
            // Increase the UI limit and rerender.
            this.setState(state => ({
                uiLimit: state.uiLimit + UI_PAGE_SIZE,
            }))
            return
        }

        // Requery with an increased max result count.
        const params = new URLSearchParams(this.props.location.search)
        let query = params.get('q') || ''

        const defaultMaxSearchResults = Math.max(this.state.resultCount || 0, 30)

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
