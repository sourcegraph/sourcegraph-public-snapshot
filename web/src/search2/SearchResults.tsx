import Loader from '@sourcegraph/icons/lib/Loader'
import ReportIcon from '@sourcegraph/icons/lib/Report'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { ReferencesGroup } from '../references/ReferencesWidget'
import { eventLogger } from '../tracking/eventLogger'
import { searchText } from './backend'
import { FileMatch } from './FileMatch'
import { parseSearchURLQuery, SearchOptions, searchOptionsEqual } from './index'
import { SearchAlert } from './SearchAlert'

interface Props {
    location: H.Location
}

interface State {
    results: GQL.IFileMatch[]
    alert: GQL.ISearchAlert | null
    loading: boolean
    searchDuration?: number
    error?: Error
    limitHit: boolean
    cloning: string[]
    missing: string[]
}

function numberWithCommas(x: any): string {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

function pluralize(str: string, n: number): string {
    return `${str}${n === 1 ? '' : 's'}`
}

export class SearchResults extends React.Component<Props, State> {
    public state: State = {
        results: [],
        alert: null,
        loading: true,
        limitHit: false,
        cloning: [],
        missing: [],
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
                        const start = Date.now()
                        return searchText(searchOptions).pipe(
                            tap(res => {
                                if (res.cloning.length > 0) {
                                    // Perform search again if there are repos still waiting to be cloned,
                                    // so we can update the results list with those repos' results.
                                    setTimeout(() => this.searchRequested.next(searchOptions), 2000)
                                }
                            }),
                            tap(
                                res =>
                                    eventLogger.log('SearchResultsFetched', {
                                        code_search: {
                                            results: {
                                                results_count: res.results.length,
                                                result_items_count: res.results.reduce(
                                                    (count, result) => count + resultItemsCount(result),
                                                    0
                                                ),
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
                                searchDuration: Date.now() - start,
                            })),
                            catchError(error => [
                                {
                                    results: [],
                                    alert: null,
                                    missing: [],
                                    cloning: [],
                                    limitHit: false,
                                    error,
                                    loading: false,
                                    searchDuration: undefined,
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
                    tap(props => {
                        const searchOptions = parseSearchURLQuery(props.location.search)
                        setTimeout(() => this.searchRequested.next(searchOptions))
                    }),
                    map(() => ({
                        results: [],
                        alert: null,
                        missing: [],
                        cloning: [],
                        limitHit: false,
                        error: undefined,
                        loading: true,
                        searchDuration: undefined,
                    }))
                )
                .subscribe(newState => this.setState(newState as State), err => console.error(err))
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let alert: {
            title: string
            description?: string | null
            proposedQueries?: GQL.ISearchQuery2Description[]
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
            alert = { title: 'No results' }
        }

        let totalMatches = 0
        let totalResults = 0
        let totalRepos = 0
        const seenRepos = new Set<string>()
        for (const result of this.state.results) {
            const parsed = new URL(result.resource)
            if (!seenRepos.has(parsed.pathname)) {
                seenRepos.add(parsed.pathname)
                totalRepos += 1
            }
            totalResults += resultItemsCount(result)
        }

        return (
            <div className="search-results2">
                {this.state.results.length > 0 && (
                    <div className="search-results2__header">
                        <div className="search-results2__badge">{numberWithCommas(totalResults)}</div>
                        <div className="search-results2__label">{pluralize('result', totalResults)} in</div>
                        <div className="search-results2__badge">{numberWithCommas(totalRepos)}</div>
                        <div className="search-results2__label">{pluralize('repo', totalRepos)} </div>
                        <div className="search-results2__duration">{this.state.searchDuration! / 1000} seconds</div>
                    </div>
                )}
                {this.state.cloning.map((repoPath, i) => (
                    <ReferencesGroup hidden={true} repoPath={repoPath} key={i} isLocal={false} icon={Loader} />
                ))}
                {this.state.missing.map((repoPath, i) => (
                    <ReferencesGroup hidden={true} repoPath={repoPath} key={i} isLocal={false} icon={ReportIcon} />
                ))}
                {this.state.loading && <Loader className="icon-inline" />}
                {alert && (
                    <SearchAlert
                        title={alert.title}
                        description={alert.description || undefined}
                        proposedQueries={this.state.alert ? this.state.alert.proposedQueries : undefined}
                        location={this.props.location}
                    />
                )}
                {this.state.results.map((result, i) => {
                    const prevTotal = totalMatches
                    totalMatches += resultItemsCount(result)
                    const expanded = prevTotal <= 500
                    return this.renderResult(i, result, expanded)
                })}
            </div>
        )
    }

    private logEvent = () => eventLogger.log('SearchResultClicked')

    private renderResult(key: number, result: GQL.SearchResult, expanded: boolean): JSX.Element | undefined {
        switch (result.__typename) {
            case 'FileMatch':
                return (
                    <FileMatch
                        key={key}
                        location={this.props.location}
                        result={result}
                        onSelect={this.logEvent}
                        expanded={expanded}
                    />
                )
        }
        return undefined
    }
}

function resultItemsCount(result: GQL.SearchResult): number {
    switch (result.__typename) {
        case 'FileMatch':
            return result.lineMatches.length
    }
    return 1
}
