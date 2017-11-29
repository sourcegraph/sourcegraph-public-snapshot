import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import * as React from 'react'
import { match } from 'react-router'
import { defer } from 'rxjs/observable/defer'
import { catchError } from 'rxjs/operators/catchError'
import { delay } from 'rxjs/operators/delay'
import { retryWhen } from 'rxjs/operators/retryWhen'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import {
    ECLONEINPROGESS,
    EREPONOTFOUND,
    EREVNOTFOUND,
    ERREPOSEEOTHER,
    fetchPhabricatorRepo,
    RepoSeeOtherError,
    resolveRev,
} from './backend'
import { parseRepoRev } from './index'
import { Repository } from './Repository'

interface Props {
    location: H.Location
    history: H.History
    match: match<{ repoRev: string; filePath?: string }>
    isDirectory: boolean
    isLightTheme: boolean
}

interface State {
    commitID?: string
    defaultBranch?: string
    phabricatorCallsign?: string
    cloneInProgress: boolean
    notFound: boolean
}

/**
 * Takes repo and rev from the matched URL route, resolves it to a commit ID and passes it to the Repository component.
 * Renders 404 if the repo or rev was not found, clone in progress while the repo is being cloned.
 */
export class RepositoryResolver extends React.Component<Props, State> {
    public state: State = { cloneInProgress: false, notFound: false }
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    switchMap(props => {
                        const repoRev = props.match.params.repoRev
                        if (!repoRev) {
                            return [undefined]
                        }
                        const { repoPath, rev } = parseRepoRev(repoRev)
                        // Defer Observable so it retries the request on resubscription
                        return (
                            defer(() => resolveRev({ repoPath, rev }))
                                // On a CloneInProgress error, retry after 1s
                                .pipe(
                                    retryWhen(errors =>
                                        errors.pipe(
                                            tap(err => {
                                                switch (err.code) {
                                                    case ERREPOSEEOTHER:
                                                        window.location.href = (err as RepoSeeOtherError).redirectURL
                                                    case ECLONEINPROGESS:
                                                        // Display cloning screen to the user and retry
                                                        this.setState({ cloneInProgress: true })
                                                        return
                                                    case EREPONOTFOUND:
                                                    case EREVNOTFOUND:
                                                        // Display 404 to the user and do not retry
                                                        this.setState({ notFound: true })
                                                }
                                                // Don't retry
                                                throw err
                                            }),
                                            delay(1000)
                                        )
                                    )
                                )
                        )
                    })
                )
                .subscribe(
                    resolvedRev => this.setState({ ...resolvedRev, cloneInProgress: false }),
                    err => console.error(err)
                )
        )
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    switchMap(props => {
                        if (!props.match.params.repoRev) {
                            return [null]
                        }
                        const { repoPath } = parseRepoRev(props.match.params.repoRev)
                        return fetchPhabricatorRepo({ repoPath }).pipe(
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    })
                )
                .subscribe(
                    phabRepo => {
                        if (phabRepo) {
                            this.setState({ phabricatorCallsign: phabRepo.callsign })
                        }
                    },
                    err => console.error(err)
                )
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.match.params.repoRev !== nextProps.match.params.repoRev) {
            // clear state so the child won't render until the revision is resolved for new props
            this.setState({
                cloneInProgress: false,
                notFound: false,
                commitID: undefined,
                phabricatorCallsign: undefined,
            })
            this.componentUpdates.next(nextProps)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { repoPath, rev } = parseRepoRev(this.props.match.params.repoRev)
        if (this.state.notFound) {
            return (
                <HeroPage
                    icon={DirectionalSignIcon}
                    title="404: Not Found"
                    subtitle="Sorry, the requested URL was not found."
                />
            )
        }
        if (this.state.cloneInProgress) {
            return (
                <HeroPage
                    icon={RepoIcon}
                    title={repoPath
                        .split('/')
                        .slice(1)
                        .join('/')}
                    subtitle="Cloning in progress"
                />
            )
        }
        if (!this.state.commitID || !this.state.defaultBranch) {
            // render empty until commit resolved
            return null
        }
        return (
            <Repository
                repoPath={repoPath}
                rev={rev}
                filePath={this.props.match.params.filePath}
                commitID={this.state.commitID}
                defaultBranch={this.state.defaultBranch}
                location={this.props.location}
                history={this.props.history}
                isLightTheme={this.props.isLightTheme}
                isDirectory={this.props.isDirectory}
                phabricatorCallsign={this.state.phabricatorCallsign}
            />
        )
    }
}
