import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import * as React from 'react'
import { match } from 'react-router'
import 'rxjs/add/observable/defer'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/delay'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/retryWhen'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from '../components/HeroPage'
import { ECLONEINPROGESS, EREPONOTFOUND, resolveRev } from './backend'
import { Repository } from './Repository'

interface Props {
    location: H.Location
    history: H.History
    match: match<{ repoRev: string, filePath?: string }>
    onToggleFullWidth: () => void
    isFullWidth: boolean
}

interface State {
    commitID?: string
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
                .switchMap(props => {
                    if (!props.match.params.repoRev) {
                        return [undefined]
                    }
                    const [repoPath, rev] = props.match.params.repoRev.split('@')
                    // Defer Observable so it retries the request on resubscription
                    return Observable.defer(() => resolveRev({ repoPath, rev }))
                        // On a CloneInProgress error, retry after 5s
                        .retryWhen(errors => errors
                            .do(err => {
                                if (err.code === ECLONEINPROGESS) {
                                    // Display cloning screen to the user and retry
                                    this.setState({ cloneInProgress: true })
                                    return
                                }
                                if (err.code === EREPONOTFOUND) {
                                    // Display 404to the user and do not retry
                                    this.setState({ notFound: true })
                                }
                                // Don't retry other errors
                                throw err
                            })
                            .delay(1000)
                        )
                        // Log other errors but don't break the stream
                        .catch(err => {
                            console.error(err)
                            return []
                        })
                })
                .subscribe(commitID => {
                    this.setState({ commitID, cloneInProgress: false })
                }, err => console.error(err))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.match.params.repoRev !== nextProps.match.params.repoRev) {
            // clear state so the child won't render until the revision is resolved for new props
            this.state = { cloneInProgress: false, notFound: false }
            this.componentUpdates.next(nextProps)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const [repoPath, rev] = this.props.match.params.repoRev.split('@')
        if (this.state.notFound) {
            return <HeroPage icon={DirectionalSignIcon} title='404: Not Found' subtitle='Sorry, the requested URL was not found.' />
        }
        if (this.state.cloneInProgress) {
            return <HeroPage icon={RepoIcon} title={repoPath.split('/').slice(1).join('/')} subtitle='Cloning in progress' />
        }
        if (this.props.match.params.repoRev && !this.state.commitID) {
            // commit not yet resolved but required if repoPath prop is provided;
            // render empty until commit resolved
            return null
        }
        return (
            <Repository
                repoPath={repoPath}
                rev={rev}
                filePath={this.props.match.params.filePath}
                commitID={this.state.commitID!}
                location={this.props.location}
                history={this.props.history}
                onToggleFullWidth={this.props.onToggleFullWidth}
                isFullWidth={this.props.isFullWidth}
            />
        )
    }
}
