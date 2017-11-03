import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import ReportIcon from '@sourcegraph/icons/lib/Report'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Link } from 'react-router-dom'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { ReferencesGroup } from '../references/ReferencesWidget'
import { events, viewEvents } from '../tracking/events'
import { searchText } from './backend'
import { buildSearchURLQuery, parseSearchURLQuery, SearchOptions, searchOptionsEqual } from './index'

interface Props {
    location: H.Location
}

interface State {
    results: GQL.IFileMatch[]
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
        loading: true,
        limitHit: false,
        cloning: [],
        missing: [],
    }

    private componentUpdates = new Subject<Props>()
    private searchRequested = new Subject<SearchOptions>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        viewEvents.SearchResults.log()

        this.subscriptions.add(
            this.searchRequested
                // Don't search using stale search options.
                .filter(searchOptions => {
                    const currentSearchOptions = parseSearchURLQuery(this.props.location.search)
                    return searchOptionsEqual(searchOptions, currentSearchOptions)
                })
                .switchMap(searchOptions => {
                    const start = Date.now()
                    return searchText(searchOptions)
                        .do(res => {
                            if (res.cloning.length > 0) {
                                // Perform search again if there are repos still waiting to be cloned,
                                // so we can update the results list with those repos' results.
                                setTimeout(() => this.searchRequested.next(searchOptions), 2000)
                            }
                        })
                        .do(res => events.SearchResultsFetched.log({
                            code_search: {
                                results: {
                                    files_count: res.results.length,
                                    matches_count: res.results.reduce((count, fileMatch) => count + fileMatch.lineMatches.length, 0),
                                },
                            },
                        }), error => {
                            events.SearchResultsFetchFailed.log({ code_search: { error_message: error.message } })
                            console.error(error)
                        })
                        .map(res => ({ ...res, error: undefined, loading: false, searchDuration: Date.now() - start }))
                        .catch(error => [{ results: [], missing: [], cloning: [], limitHit: false, error, loading: false, searchDuration: undefined }])
                })
                .subscribe(
                newState => this.setState(newState as State),
                err => console.error(err)
                )
        )

        this.subscriptions.add(
            this.componentUpdates
                .startWith(this.props)
                .do(props => {
                    const searchOptions = parseSearchURLQuery(props.location.search)
                    setTimeout(() => this.searchRequested.next(searchOptions))
                })
                .map(() => ({ results: [], missing: [], cloning: [], limitHit: false, error: undefined, loading: true, searchDuration: undefined }))
                .subscribe(
                newState => this.setState(newState as State),
                err => console.error(err)
                )
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {

        let alertTitle: string | JSX.Element | undefined
        let alertDetails: string | JSX.Element | undefined
        if (this.state.error) {
            if (this.state.error.message.includes('no query terms or regexp specified')) {
                alertTitle = ''
                alertDetails = 'Enter terms to search...'
            } else if (this.state.error.message.includes('no repositories included')) {
                alertTitle = 'No repositories matched by current filters'
                alertDetails = querySuggestionForAllReposExcluded(parseSearchURLQuery(this.props.location.search))
            } else {
                alertTitle = 'Something went wrong!'
                alertDetails = upperFirst(this.state.error.message)
            }
        } else if (this.state.loading) {
            alertTitle = <Loader className='icon-inline' />
        } else if (this.state.results.length === 0 && this.state.missing.length === 0 && this.state.cloning.length === 0) {
            alertTitle = 'No results'
        }

        let totalMatches = 0
        let totalResults = 0
        let totalFiles = 0
        let totalRepos = 0
        const seenRepos = new Set<string>()
        for (const result of this.state.results) {
            const parsed = new URL(result.resource)
            if (!seenRepos.has(parsed.pathname)) {
                seenRepos.add(parsed.pathname)
                totalRepos += 1
            }
            totalFiles += 1
            totalResults += result.lineMatches.length
        }

        const logEvent = () => events.SearchResultClicked.log()

        const searchOptions = parseSearchURLQuery(this.props.location.search)

        return (
            <div className='search-results2'>
                {
                    this.state.results.length > 0 &&
                    <div className='search-results2__header'>
                        <div className='search-results2__badge'>{numberWithCommas(totalResults)}</div>
                        <div className='search-results2__label'>{pluralize('result', totalResults)} in</div>
                        <div className='search-results2__badge'>{numberWithCommas(totalFiles)}</div>
                        <div className='search-results2__label'>{pluralize('file', totalFiles)}  in</div>
                        <div className='search-results2__badge'>{numberWithCommas(totalRepos)}</div>
                        <div className='search-results2__label'>{pluralize('repo', totalRepos)} </div>
                        <div className='search-results2__duration'>{this.state.searchDuration! / 1000} seconds</div>
                    </div>
                }
                {
                    this.state.cloning.map((repoPath, i) =>
                        <ReferencesGroup hidden={true} repoPath={repoPath} key={i} isLocal={false} icon={Loader} />
                    )
                }
                {
                    this.state.missing.map((repoPath, i) =>
                        <ReferencesGroup hidden={true} repoPath={repoPath} key={i} isLocal={false} icon={ReportIcon} />
                    )
                }
                {
                    (alertTitle || alertDetails) &&
                    <div className='search-results2__alert'>
                        {alertTitle && <h1 className='search-results2__alert-title'>{alertTitle}</h1>}
                        {alertDetails && <p className='search-results2__alert-details'>{alertDetails}</p>}
                    </div>
                }
                {
                    this.state.results.map((result, i) => {
                        const prevTotal = totalMatches
                        totalMatches += result.lineMatches.length
                        const parsed = new URL(result.resource)
                        const repoPath = parsed.hostname + parsed.pathname
                        const rev = parsed.search.substr('?'.length)
                        const filePath = parsed.hash.substr('#'.length)
                        const refs = result.lineMatches.map(match => ({
                            range: {
                                start: {
                                    character: match.offsetAndLengths[0][0],
                                    line: match.lineNumber,
                                },
                                end: {
                                    character: match.offsetAndLengths[0][0] + match.offsetAndLengths[0][1],
                                    line: match.lineNumber,
                                },
                            },
                            uri: result.resource,
                            repoURI: repoPath,
                        }))

                        return <ReferencesGroup
                            hidden={prevTotal > 500}
                            repoPath={repoPath}
                            localRev={rev}
                            filePath={filePath}
                            key={i}
                            refs={refs}
                            isLocal={false}
                            icon={RepoIcon}
                            onSelect={logEvent}
                            searchOptions={searchOptions}
                        />
                    })
                }
            </div>
        )
    }
}

function querySuggestionForAllReposExcluded(options: SearchOptions): string | JSX.Element | undefined {
    const omitQueryFields = (query: string, omitField: string): string => query.split(' ').filter(token => !token.startsWith(omitField + ':')).join(' ')

    let suggestion: SearchOptions | undefined
    let reason: string | JSX.Element | undefined
    if (!suggestion && options.query.includes('repogroup:') || options.scopeQuery.includes('repogroup:')) {
        suggestion = {
            query: omitQueryFields(options.query, 'repogroup'),
            scopeQuery: omitQueryFields(options.scopeQuery, 'repogroup'),
        }
        reason = 'omitting the repository group filter'
        if (options.scopeQuery.includes('repogroup:')) {
            reason += ' in the search scope'
        }
    }

    const repoFields = (options.query + ' ' + options.scopeQuery).split(' ').filter(token => token.startsWith('repo:'))
    if (!suggestion && repoFields.length > 1) {
        // Suggest union'ing multiple repo: field values, in case the user thought separate
        // fields were OR'd not AND'd.
        const values = repoFields.map(token => token.replace(/^repo:/, '')).filter(s => !!s)
        suggestion = {
            query: omitQueryFields(options.query, 'repo') + ` repo:${values.join('|')}`,
            scopeQuery: omitQueryFields(options.scopeQuery, 'repo'),
        }
        reason = 'using a single pattern instead of multiple intersecting repo: filters'
    }

    if (!suggestion && repoFields.length === 1) {
        const value = repoFields[0].replace(/^repo:/, '')
        reason = (
            <span>
                The repo: filter value <code>{value}</code> matched no repositories.
                Check that it is a valid regular expression that matches repository paths
                such as <code>github.com/foo/bar</code> and <code>example.com/foo</code>.
                For example, the following filter would match both: <code>repo:foo</code>.
            </span>
        )
    }

    if (suggestion) {
        const url = '?' + buildSearchURLQuery(suggestion)
        return (
            <span>Did you mean: <Link className='search-results2__query-fix' to={url}>{suggestion.scopeQuery} {suggestion.query}</Link> {reason ? `(${reason})` : ''}</span>
        )
    }
    if (reason) {
        return reason
    }

    return undefined
}
