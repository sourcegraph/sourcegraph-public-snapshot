import ArrowCollapseVerticalIcon from '@sourcegraph/icons/lib/ArrowCollapseVertical'
import ArrowExpandVerticalIcon from '@sourcegraph/icons/lib/ArrowExpandVertical'
import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import DocumentIcon from '@sourcegraph/icons/lib/Document'
import HourglassIcon from '@sourcegraph/icons/lib/Hourglass'
import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import ReportIcon from '@sourcegraph/icons/lib/Report'
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
import { RepoSearchResult } from './RepoSearchResult'
import { RepositorySearchResult } from './RepositorySearchResult'
import { SavedQueryCreateForm } from './SavedQueryCreateForm'
import { SearchAlert } from './SearchAlert'

interface Props {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onFilterChosen: (value: string) => void
    navbarSearchQuery: string
}

interface State {
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
    showModal?: boolean
    didSave?: boolean
    user?: GQL.IUser | null
    dynamicFilters: GQL.ISearchFilter[]
    allExpanded?: boolean
}

const ALL_EXPANDED_LOCAL_STORAGE_KEY = 'allExpanded'
const DATA_CENTER_UPGRADE_STRING =
    'Upgrade to Sourcegraph Data Center for distributed on-the-fly search and near-instant indexed search.'
const SEARCH_TIMED_OUT_DEFAULT_TITLE = 'Search timed out'

export class SearchResults extends React.Component<Props, State> {
    private static SHOW_MISSING = true

    public state: State = {
        results: [],
        alert: null,
        loading: true,
        limitHit: false,
        cloning: [],
        missing: [],
        timedout: [],
        didSave: false,
        showModal: false,
        dynamicFilters: [],
        allExpanded: localStorage.getItem(ALL_EXPANDED_LOCAL_STORAGE_KEY) === 'true',
    }

    private componentUpdates = new Subject<Props>()
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
                            })),
                            catchError(error => [
                                {
                                    results: [],
                                    resultCount: undefined,
                                    approximateResultCount: undefined,
                                    alert: null,
                                    elapsedMilliseconds: undefined,
                                    missing: [],
                                    cloning: [],
                                    limitHit: false,
                                    error,
                                    loading: false,
                                    didSave: false,
                                    showModal: false,
                                    dynamicFilters: [],
                                    allExpanded: false,
                                },
                            ])
                        )
                    })
                )
                .subscribe(newState => this.setState(newState as State), err => console.error(err))
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
                        limitHit: false,
                        error: undefined,
                        loading: true,
                        didSave: false,
                        showModal: false,
                        dynamicFilters: [],
                        allExpanded: localStorage.getItem(ALL_EXPANDED_LOCAL_STORAGE_KEY) === 'true',
                    }))
                )
                .subscribe(newState => this.setState(newState as State), err => console.error(err))
        )
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
    }

    public componentWillReceiveProps(newProps: Props): void {
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

    public render(): JSX.Element | null {
        let alert: {
            title: string
            description?: string | null
            proposedQueries?: GQL.ISearchQueryDescription[]
            errorBody?: React.ReactFragment
        } | null = null
        if (this.state.error) {
            if (this.state.error.message.includes('no query terms or regexp specified')) {
                alert = { title: '', description: 'Enter terms to search...' }
            } else {
                alert = { title: 'Something went wrong', description: upperFirst(this.state.error.message) }
            }
        } else if (this.state.alert) {
            alert = this.state.alert
        } else if (
            !this.state.loading &&
            this.state.results.length === 0 &&
            this.state.missing.length === 0 &&
            this.state.cloning.length === 0
        ) {
            const searchTimeoutParameterEnabled = window.context.experimentalFeatures.searchTimeoutParameterEnabled
            const defaultTimeoutAlert = {
                title: SEARCH_TIMED_OUT_DEFAULT_TITLE,
                description: searchTimeoutParameterEnabled
                    ? "Try narrowing your query, or specifying a longer 'timeout:' in your query."
                    : 'Try narrowing your query.',
            }
            const longerTimeoutString = searchTimeoutParameterEnabled
                ? "Specify a longer 'timeout:' in your query."
                : ''
            if (this.state.timedout.length > 0) {
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

        return (
            <div className="search-results">
                {this.state.results.length > 0 && (
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
                <div className="search-results__list">
                    {this.state.showModal && (
                        <ModalContainer
                            onClose={this.onModalClose}
                            component={
                                <SavedQueryCreateForm
                                    user={this.props.user}
                                    values={{ query: parsedQuery ? parsedQuery.query : '' }}
                                    onDidCancel={this.onModalClose}
                                    onDidCreate={this.onDidCreateSavedQuery}
                                />
                            }
                        />
                    )}
                    <div className="search-results__info">
                        {(this.state.timedout.length > 0 ||
                            this.state.cloning.length > 0 ||
                            this.state.results.length > 0) && (
                            <small className="search-results__info-row">
                                <div className="search-results__info-row-left">
                                    {(this.state.timedout.length > 0 || this.state.cloning.length > 0) && (
                                        <span className="search-results__info-notice">
                                            <HourglassIcon className="icon-inline" />
                                            {this.state.timedout.length > 0 && (
                                                <span data-tooltip={this.state.timedout.join('\n')}>
                                                    {this.state.timedout.length}&nbsp;
                                                    {pluralize(
                                                        'repository',
                                                        this.state.timedout.length,
                                                        'repositories'
                                                    )}{' '}
                                                    timed out
                                                </span>
                                            )}
                                            {this.state.timedout.length > 0 &&
                                                this.state.cloning.length > 0 && <span>&nbsp;and&nbsp;</span>}
                                            {this.state.cloning.length > 0 && (
                                                <span data-tooltip={this.state.cloning.join('\n')}>
                                                    {this.state.cloning.length}&nbsp;
                                                    {pluralize(
                                                        'repository',
                                                        this.state.cloning.length,
                                                        'repositories'
                                                    )}{' '}
                                                    cloning
                                                </span>
                                            )}
                                            &nbsp;(reload to try again
                                            {window.context.experimentalFeatures.searchTimeoutParameterEnabled &&
                                                ", or specify a longer 'timeout:' in your query"}
                                            )
                                        </span>
                                    )}
                                    {typeof this.state.approximateResultCount === 'string' &&
                                        typeof this.state.resultCount === 'number' && (
                                            <span className="search-results__stats">
                                                {this.state.approximateResultCount}{' '}
                                                {pluralize('result', this.state.resultCount)}
                                                {typeof this.state.elapsedMilliseconds === 'number' && (
                                                    <>
                                                        {' '}
                                                        in {(this.state.elapsedMilliseconds / 1000).toFixed(2)} seconds
                                                    </>
                                                )}
                                            </span>
                                        )}
                                </div>
                                <div className="search-results__info-row-right">
                                    {this.state.allExpanded ? (
                                        <button onClick={this.expandAllResults} className="btn btn-link">
                                            <ArrowCollapseVerticalIcon
                                                className="icon-inline"
                                                data-tooltip="Collapse"
                                            />
                                        </button>
                                    ) : (
                                        <button onClick={this.expandAllResults} className="btn btn-link">
                                            <ArrowExpandVerticalIcon className="icon-inline" data-tooltip="Expand" />
                                        </button>
                                    )}
                                    {!this.state.didSave &&
                                        this.state.user && (
                                            <button onClick={this.showSaveQueryModal} className="btn btn-link">
                                                <SaveIcon className="icon-inline" /> Save this search query
                                            </button>
                                        )}
                                    {this.state.didSave && (
                                        <span>
                                            <CheckmarkIcon className="icon-inline" /> Query saved
                                        </span>
                                    )}
                                </div>
                            </small>
                        )}
                        {!this.state.alert &&
                            !this.state.error &&
                            !this.state.loading &&
                            showDotComMarketing && <ServerBanner />}
                    </div>
                    {this.state.results.length > 0}
                    {SearchResults.SHOW_MISSING &&
                        this.state.missing.map((repoPath, i) => (
                            <RepoSearchResult repoPath={repoPath} key={i} icon={ReportIcon} />
                        ))}
                    {this.state.loading && <Loader className="icon-inline" />}
                    {this.state.results.slice(0, 75).map((result, i) => this.renderResult(i, result, i <= 15))}
                    {this.state.limitHit && (
                        <button className="btn btn-link search-results__more" onClick={this.showMoreResults}>
                            Show more
                        </button>
                    )}
                    {alert && (
                        <SearchAlert
                            className="search-results__alert"
                            title={alert.title}
                            description={alert.description || undefined}
                            proposedQueries={this.state.alert ? this.state.alert.proposedQueries : undefined}
                            location={this.props.location}
                            errorBody={alert.errorBody}
                        />
                    )}
                </div>
            </div>
        )
    }

    private logEvent = () => eventLogger.log('SearchResultClicked')

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
                        allExpanded={this.state.allExpanded}
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
                        allExpanded={this.state.allExpanded}
                    />
                )
        }
        return undefined
    }

    private showMoreResults = () => {
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
