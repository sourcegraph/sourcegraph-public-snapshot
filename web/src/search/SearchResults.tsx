import Loader from '@sourcegraph/icons/lib/Loader'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import ReportIcon from '@sourcegraph/icons/lib/Report'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { timer } from 'rxjs/observable/timer'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { multicast } from 'rxjs/operators/multicast'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { take } from 'rxjs/operators/take'
import { takeWhile } from 'rxjs/operators/takeWhile'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { ReferencesGroup } from '../references/ReferencesWidget'
import { eventLogger } from '../tracking/eventLogger'
import { searchText } from './backend'
import { parseSearchURLQuery } from './index'

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
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SearchResults')
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    switchMap(props => {
                        const start = Date.now()
                        const searchOptions = parseSearchURLQuery(props.location.search)
                        // Repeat the search every 1s until no more repos are clone in progress
                        return timer(0, 1000).pipe(
                            mergeMap(() => searchText(searchOptions)),
                            // Make sure that last item is still included
                            // Can't use takeWhile here because it would omit the last element (the first which the predicate returns false for)
                            multicast<GQL.ISearchResults>(
                                () => new ReplaySubject<GQL.ISearchResults>(1),
                                textSearch =>
                                    textSearch.pipe(
                                        takeWhile(res => res.cloning.length > 0),
                                        concat(textSearch.pipe(take(1)))
                                    )
                            ),
                            map(res => ({
                                ...res,
                                error: undefined,
                                loading: false,
                                searchDuration: Date.now() - start,
                            })),
                            catchError(error => {
                                console.error(error)
                                return [
                                    {
                                        results: [],
                                        missing: [],
                                        cloning: [],
                                        limitHit: false,
                                        error,
                                        loading: false,
                                        searchDuration: undefined,
                                    },
                                ]
                            }),
                            // Reset to loading state
                            startWith({
                                results: [],
                                missing: [],
                                cloning: [],
                                limitHit: false,
                                error: undefined,
                                loading: true,
                                searchDuration: undefined,
                            })
                        )
                    })
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
        let alertTitle: string | JSX.Element | undefined
        let alertDetails: string | undefined
        if (this.state.error) {
            alertTitle = 'Something went wrong!'
            alertDetails = upperFirst(this.state.error.message)
        } else if (this.state.loading) {
            alertTitle = <Loader className="icon-inline" />
        } else if (
            this.state.results.length === 0 &&
            this.state.missing.length === 0 &&
            this.state.cloning.length === 0
        ) {
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

        const logEvent = () => eventLogger.log('SearchResultClicked')

        return (
            <div className="search-results">
                {this.state.results.length > 0 && (
                    <div className="search-results__header">
                        <div className="search-results__badge">{numberWithCommas(totalResults)}</div>
                        <div className="search-results__label">{pluralize('result', totalResults)} in</div>
                        <div className="search-results__badge">{numberWithCommas(totalFiles)}</div>
                        <div className="search-results__label">{pluralize('file', totalFiles)} in</div>
                        <div className="search-results__badge">{numberWithCommas(totalRepos)}</div>
                        <div className="search-results__label">{pluralize('repo', totalRepos)} </div>
                        <div className="search-results__duration">{this.state.searchDuration! / 1000} seconds</div>
                    </div>
                )}
                {this.state.cloning.map((repoPath, i) => (
                    <ReferencesGroup hidden={true} repoPath={repoPath} key={i} isLocal={false} icon={Loader} />
                ))}
                {this.state.missing.map((repoPath, i) => (
                    <ReferencesGroup hidden={true} repoPath={repoPath} key={i} isLocal={false} icon={ReportIcon} />
                ))}
                {(alertTitle || alertDetails) && (
                    <div className="search-results__alert">
                        {alertTitle && <h1 className="search-results__alert-title">{alertTitle}</h1>}
                        {alertDetails && <p className="search-results__alert-details">{alertDetails}</p>}
                    </div>
                )}
                {this.state.results.map((result, i) => {
                    const prevTotal = totalMatches
                    totalMatches += result.lineMatches.length
                    const parsed = new URL(result.resource)
                    const repoPath = parsed.hostname + parsed.pathname
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

                    return (
                        <ReferencesGroup
                            hidden={prevTotal > 500}
                            repoPath={repoPath}
                            filePath={filePath}
                            key={i}
                            refs={refs}
                            isLocal={false}
                            icon={RepoIcon}
                            onSelect={logEvent}
                        />
                    )
                })}
            </div>
        )
    }
}
