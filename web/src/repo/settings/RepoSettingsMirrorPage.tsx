import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { interval } from 'rxjs/observable/interval'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { checkMirrorRepositoryConnection, updateMirrorRepository } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepository } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'

interface UpdateMirrorRepositoryActionContainerProps {
    repo: GQL.IRepository
    onDidUpdateRepository: () => void
    disabled: boolean
    disabledReason: string | undefined
}

class UpdateMirrorRepositoryActionContainer extends React.PureComponent<UpdateMirrorRepositoryActionContainerProps> {
    private componentUpdates = new Subject<UpdateMirrorRepositoryActionContainerProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // If repository is cloning, poll until it's done.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => props.repo.mirrorInfo.cloneInProgress),
                    distinctUntilChanged(),
                    filter(cloneInProgress => cloneInProgress),
                    switchMap(() =>
                        interval(3000).pipe(
                            takeUntil(
                                this.componentUpdates.pipe(filter(props => !props.repo.mirrorInfo.cloneInProgress))
                            )
                        )
                    )
                )
                .subscribe(repo => this.props.onDidUpdateRepository())
        )
    }

    public componentWillReceiveProps(props: UpdateMirrorRepositoryActionContainerProps): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let title: React.ReactFragment
        let description: React.ReactFragment
        let buttonLabel: React.ReactFragment
        let buttonDisabled = false
        if (this.props.repo.mirrorInfo.cloneInProgress) {
            title = 'Cloning in progress...'
            description = 'This repository is currently being cloned from its remote repository.'
            buttonLabel = (
                <span>
                    <LoaderIcon className="icon-inline" /> Cloning...
                </span>
            )
            buttonDisabled = true
        } else if (this.props.repo.mirrorInfo.cloned) {
            title = (
                <>
                    Last refreshed:{' '}
                    {this.props.repo.mirrorInfo.updatedAt ? (
                        <Timestamp date={this.props.repo.mirrorInfo.updatedAt} />
                    ) : (
                        'unknown'
                    )}{' '}
                </>
            )
            description =
                'This repository is automatically updated from its remote repository periodically and when accessed by a user.'
            buttonLabel = 'Refresh now'
        } else {
            title = 'Clone this repository'
            description = 'This repository has not yet been cloned from its remote repository.'
            buttonLabel = 'Clone now'
        }

        return (
            <ActionContainer
                title={title}
                description={<div>{description}</div>}
                buttonLabel={buttonLabel}
                buttonDisabled={buttonDisabled || this.props.disabled}
                buttonSubtitle={this.props.disabledReason}
                flashText="Added to queue"
                run={this.updateMirrorRepository}
            />
        )
    }

    private updateMirrorRepository = () => {
        const p = updateMirrorRepository({ repository: this.props.repo.id })
            .toPromise()
            .then(result => this.props.onDidUpdateRepository())
        return p
    }
}

interface CheckMirrorRepositoryConnectionActionContainerProps {
    repo: GQL.IRepository
    onDidUpdateReachability: (reachable: boolean | undefined) => void
}

interface CheckMirrorRepositoryConnectionActionContainerState {
    loading: boolean
    result?: GQL.ICheckMirrorRepositoryConnectionResult | null
    errorDescription?: string
}

class CheckMirrorRepositoryConnectionActionContainer extends React.PureComponent<
    CheckMirrorRepositoryConnectionActionContainerProps,
    CheckMirrorRepositoryConnectionActionContainerState
> {
    public state: CheckMirrorRepositoryConnectionActionContainerState = { loading: false }

    private checkRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.checkRequests
                .pipe(
                    tap(() => {
                        this.setState({ errorDescription: undefined, result: undefined, loading: true })
                        this.props.onDidUpdateReachability(undefined)
                    }),
                    switchMap(() =>
                        checkMirrorRepositoryConnection({ repository: this.props.repo.id }).pipe(
                            catchError(error => {
                                this.setState({ errorDescription: error.message, result: undefined, loading: false })
                                this.props.onDidUpdateReachability(false)
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    result => {
                        this.setState({ result, loading: false })
                        this.props.onDidUpdateReachability(result.error === null)
                    },
                    error => console.log(error)
                )
        )

        // Run the check upon initial mount, so the user sees the information without needing to click.
        this.checkRequests.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <BaseActionContainer
                title="Check connection to remote repository"
                description={<span>Diagnose problems cloning or updating from the remote repository.</span>}
                action={
                    <button
                        className="btn btn-primary"
                        disabled={this.state.loading}
                        onClick={this.checkMirrorRepositoryConnection}
                    >
                        Check connection
                    </button>
                }
                details={
                    <>
                        {this.state.errorDescription && (
                            <div className="alert alert-danger action-container__alert">
                                Error: {this.state.errorDescription}
                            </div>
                        )}
                        {this.state.loading && (
                            <div className="alert alert-primary action-container__alert">
                                <LoaderIcon className="icon-inline" /> Checking connection...
                            </div>
                        )}
                        {this.state.result &&
                            (this.state.result.error === null ? (
                                <div className="alert alert-success action-container__alert">
                                    <CheckmarkIcon className="icon-inline" /> The remote repository is reachable.
                                </div>
                            ) : (
                                <div className="alert alert-danger action-container__alert">
                                    <p>The remote repository is unreachable. Logs follow.</p>
                                    <div>
                                        <pre className="check-mirror-repository-connection-action-container__log">
                                            <code>{this.state.result.error}</code>
                                        </pre>
                                    </div>
                                </div>
                            ))}
                    </>
                }
            />
        )
    }

    private checkMirrorRepositoryConnection = () => this.checkRequests.next()
}

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
    user: GQL.IUser
    onDidUpdateRepository: (update: Partial<GQL.IRepository>) => void
}

interface State {
    /**
     * The repository object, refreshed after we make changes that modify it.
     */
    repo: GQL.IRepository

    /**
     * Whether the repository connection check reports that the repository is reachable.
     */
    reachable?: boolean

    loading: boolean
    error?: string
}

/**
 * The repository settings mirror page.
 */
export class RepoSettingsMirrorPage extends React.PureComponent<Props, State> {
    private repoUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsMirror')

        this.subscriptions.add(
            this.repoUpdates
                .pipe(switchMap(() => fetchRepository(this.props.repo.uri)))
                .subscribe(repo => this.setState({ repo }), err => this.setState({ error: err.message }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-settings-mirror-page">
                <PageTitle title="Mirror settings" />
                <h2>Mirroring and cloning</h2>
                {this.state.loading && <LoaderIcon className="icon-inline" />}
                {this.state.error && <div className="alert alert-danger">{this.state.error}</div>}
                <dl>
                    <dt>
                        Remote repository URL{' '}
                        <div className="settings-page__form-label-badge">
                            <LockIcon className="icon-inline" /> Only visible to site admins
                        </div>
                    </dt>
                    <dd>
                        <div className="form-control settings-page__form-fake-input">
                            <code className="settings-page__form-fake-input--code">
                                {this.props.repo.mirrorInfo.remoteURL || '(unknown)'}
                            </code>
                        </div>
                    </dd>
                    <p className="settings-page__form-notice">
                        <small>
                            Configure repository mirroring in{' '}
                            {this.state.repo.viewerCanAdminister ? (
                                <Link to="/site-admin/configuration">site configuration</Link>
                            ) : (
                                'site configuration'
                            )}.
                        </small>
                    </p>
                </dl>
                {
                    <UpdateMirrorRepositoryActionContainer
                        repo={this.state.repo}
                        onDidUpdateRepository={this.onDidUpdateRepository}
                        disabled={typeof this.state.reachable === 'boolean' && !this.state.reachable}
                        disabledReason={
                            typeof this.state.reachable === 'boolean' && !this.state.reachable
                                ? 'Not reachable'
                                : undefined
                        }
                    />
                }
                {
                    <CheckMirrorRepositoryConnectionActionContainer
                        repo={this.state.repo}
                        onDidUpdateReachability={this.onDidUpdateReachability}
                    />
                }
                {typeof this.state.reachable === 'boolean' &&
                    !this.state.reachable && (
                        <div className="alert alert-info repo-settings-mirror-page__troubleshooting">
                            Problems cloning or updating this repository?
                            <ul className="repo-settings-mirror-page__steps">
                                <li className="repo-settings-mirror-page__step">
                                    Inspect the <strong>Check connection</strong> error log output to see why the remote
                                    repository is not reachable.
                                </li>
                                <li className="repo-settings-mirror-page__step">
                                    <code>
                                        <strong>No ECDSA host key is known ... Host key verification failed?</strong>
                                    </code>{' '}
                                    See{' '}
                                    <a href="https://about.sourcegraph.com/docs/server/config/repositories#ssh-authentication-config-keys-known_hosts">
                                        SSH repository authentication documentation
                                    </a>{' '}
                                    for how to provide an SSH <code>known_hosts</code> file with the remote host's SSH
                                    host key.
                                </li>
                                <li className="repo-settings-mirror-page__step">
                                    Consult{' '}
                                    <a href="https://about.sourcegraph.com/docs/server/config/repositories">
                                        Sourcegraph Server repositories documentation
                                    </a>{' '}
                                    for resolving other authentication issues (such as HTTPS certificates and SSH keys).
                                </li>
                                <li className="repo-settings-mirror-page__step">
                                    <a href="mailto:support@sourcegraph.com">Contact support</a> for further help.
                                </li>
                            </ul>
                        </div>
                    )}
            </div>
        )
    }

    private onDidUpdateRepository = () => {
        this.repoUpdates.next()
        this.props.onDidUpdateRepository({})
    }

    private onDidUpdateReachability = (reachable: boolean | undefined) => this.setState({ reachable })
}
