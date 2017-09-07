import * as React from 'react'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { SearchResult, searchText } from 'sourcegraph/backend'
import { ReferencesGroup } from 'sourcegraph/references/ReferencesWidget'
import { parseSearchURLQuery } from 'sourcegraph/search'
import { ParsedRouteProps } from 'sourcegraph/util/routes'

interface Props extends ParsedRouteProps { }

interface State {
    results: SearchResult[]
    loading: boolean
    searchDuration?: number
}

function numberWithCommas(x: any): string {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

function pluralize(str: string, n: number): string {
    return `${str}${n === 1 ? '' : 's'}`
}

export class SearchResults extends React.Component<Props, State> {

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            results: [],
            loading: true
        }
        this.subscriptions.add(
            this.componentUpdates
                .startWith(props)
                .switchMap(props => {
                    const start = Date.now()
                    const searchOptions = parseSearchURLQuery(props.location.search)
                    return searchText(searchOptions)
                        .catch(err => {
                            // TODO display error
                            console.error(err)
                            return []
                        })
                        .map((res: GQL.ISearchResults): State => ({ results: res.results, loading: false, searchDuration: Date.now() - start }))
                })
                .subscribe(
                    newState => this.setState(newState),
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
        if (this.state.loading) {
            return (
                <div className='searchResults'>
                    <div className='search-results__header'>
                        Working...
                    </div>
                </div>
            )
        }
        if (!this.state.results || this.state.results.length === 0) {
            return (
                <div className='searchResults'>
                    <div className='search-results__header'>
                        No results
                    </div>
                </div>
            )
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
        return (
            <div className='search-results'>
                <div className='search-results__header'>
                    <div className='search-results__badge'>{numberWithCommas(totalResults)}</div>
                    <div className='search-results__label'>{pluralize('result', totalResults)} in</div>
                    <div className='search-results__badge'>{numberWithCommas(totalFiles)}</div>
                    <div className='search-results__label'>{pluralize('file', totalFiles)}  in</div>
                    <div className='search-results__badge'>{numberWithCommas(totalRepos)}</div>
                    <div className='search-results__label'>{pluralize('repo', totalRepos)} </div>
                    <div className='search-results__duration'>{this.state.searchDuration! / 1000} seconds</div>
                </div>
                {
                    this.state.results.map((result, i) => {
                        const prevTotal = totalMatches
                        totalMatches += result.lineMatches.length
                        const parsed = new URL(result.resource)
                        const repoPath = parsed.pathname.substr('//'.length)
                        const filePath = parsed.hash.substr('#'.length)
                        const refs = result.lineMatches.map(match => ({
                            range: {
                                start: {
                                    character: match.offsetAndLengths[0][0],
                                    line: match.lineNumber
                                },
                                end: {
                                    character: match.offsetAndLengths[0][0] + match.offsetAndLengths[0][1],
                                    line: match.lineNumber
                                }
                            },
                            uri: result.resource,
                            repoURI: repoPath
                        }))

                        return <ReferencesGroup hidden={prevTotal > 500} uri={repoPath} path={filePath} key={i} refs={refs} isLocal={false} />
                    })
                }
            </div>
        )
    }
}
