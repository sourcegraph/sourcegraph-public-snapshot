import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { interval } from 'rxjs/observable/interval'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { checkMirrorRepositoryConnection, updateMirrorRepository } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepository } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'

interface MirrorRepositoryActionContainerProps {
    repo: GQL.IRepository
    onDidUpdateRepository: () => void
}

class UpdateMirrorRepositoryActionContainer extends React.PureComponent<MirrorRepositoryActionContainerProps> {
    private componentUpdates = new Subject<MirrorRepositoryActionContainerProps>()
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

    public componentWillReceiveProps(props: MirrorRepositoryActionContainerProps): void {
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
            buttonLabel = 'foo' // TODO!(sqs)
        }

        return (
            <ActionContainer
                title={title}
                description={<div>{description}</div>}
                buttonLabel={buttonLabel}
                buttonDisabled={buttonDisabled}
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

interface CheckMirrorRepositoryConnectionActionContainerState {
    loading: boolean
    result?: GQL.ICheckMirrorRepositoryConnectionResult | null
    errorDescription?: string
}

class CheckMirrorRepositoryConnectionActionContainer extends React.PureComponent<
    MirrorRepositoryActionContainerProps,
    CheckMirrorRepositoryConnectionActionContainerState
> {
    public state: CheckMirrorRepositoryConnectionActionContainerState = { loading: false }

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
                                {this.state.errorDescription}
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

    private checkMirrorRepositoryConnection = () => {
        this.setState({ errorDescription: undefined, result: undefined, loading: true })
        const p = checkMirrorRepositoryConnection({ repository: this.props.repo.id })
            .toPromise()
            .then(
                result => this.setState({ result, loading: false }),
                err => this.setState({ errorDescription: err.message, result: undefined, loading: false })
            )
        return p
    }
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
                            <code>{this.props.repo.mirrorInfo.remoteURL}</code>
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
                    />
                }
                {
                    <CheckMirrorRepositoryConnectionActionContainer
                        repo={this.state.repo}
                        onDidUpdateRepository={this.onDidUpdateRepository}
                    />
                }
            </div>
        )
    }

    private onDidUpdateRepository = () => {
        this.repoUpdates.next()
        this.props.onDidUpdateRepository({})
    }
}
