import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import LockIcon from 'mdi-react/LockIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { interval, Subject, Subscription } from 'rxjs'
import { catchError, switchMap, tap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { FeedbackText } from '../../components/FeedbackText'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { checkMirrorRepositoryConnection, updateMirrorRepository } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { DirectImportRepoAlert } from '../DirectImportRepoAlert'
import { fetchRepository } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'
import { ErrorAlert } from '../../components/alerts'

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
        that.subscriptions.add(
            interval(3000).subscribe(() => {
                that.props.onDidUpdateRepository()
            })
        )
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let title: React.ReactFragment
        let description: React.ReactFragment
        let buttonLabel: React.ReactFragment
        let buttonDisabled = false
        let info: React.ReactNode
        if (that.props.repo.mirrorInfo.cloneInProgress) {
            title = 'Cloning in progress...'
            description =
                <code>{that.props.repo.mirrorInfo.cloneProgress}</code> ||
                'This repository is currently being cloned from its remote repository.'
            buttonLabel = (
                <span>
                    <LoadingSpinner className="icon-inline" /> Cloning...
                </span>
            )
            buttonDisabled = true
            info = <DirectImportRepoAlert className="action-container__alert" />
        } else if (that.props.repo.mirrorInfo.cloned) {
            const updateSchedule = that.props.repo.mirrorInfo.updateSchedule
            title = (
                <>
                    <div>
                        Last refreshed:{' '}
                        {that.props.repo.mirrorInfo.updatedAt ? (
                            <Timestamp date={that.props.repo.mirrorInfo.updatedAt} />
                        ) : (
                            'unknown'
                        )}{' '}
                    </div>
                    {updateSchedule && (
                        <div>
                            Next scheduled update <Timestamp date={updateSchedule.due} /> (position{' '}
                            {updateSchedule.index + 1} out of {updateSchedule.total} in the schedule)
                        </div>
                    )}
                    {that.props.repo.mirrorInfo.updateQueue && !that.props.repo.mirrorInfo.updateQueue.updating && (
                        <div>
                            Queued for update (position {that.props.repo.mirrorInfo.updateQueue.index + 1} out of{' '}
                            {that.props.repo.mirrorInfo.updateQueue.total} in the queue)
                        </div>
                    )}
                </>
            )
            if (!updateSchedule) {
                description = 'This repository is automatically updated when accessed by a user.'
            } else {
                description =
                    'This repository is automatically updated from its remote repository periodically and when accessed by a user.'
            }
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
                buttonDisabled={buttonDisabled || that.props.disabled}
                buttonSubtitle={that.props.disabledReason}
                flashText="Added to queue"
                info={info}
                run={that.updateMirrorRepository}
            />
        )
    }

    private updateMirrorRepository = async (): Promise<void> => {
        await updateMirrorRepository({ repository: that.props.repo.id }).toPromise()
        that.props.onDidUpdateRepository()
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
        that.subscriptions.add(
            that.checkRequests
                .pipe(
                    tap(() => {
                        that.setState({ errorDescription: undefined, result: undefined, loading: true })
                        that.props.onDidUpdateReachability(undefined)
                    }),
                    switchMap(() =>
                        checkMirrorRepositoryConnection({ repository: that.props.repo.id }).pipe(
                            catchError(error => {
                                that.setState({ errorDescription: error.message, result: undefined, loading: false })
                                that.props.onDidUpdateReachability(false)
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    result => {
                        that.setState({ result, loading: false })
                        that.props.onDidUpdateReachability(result.error === null)
                    },
                    error => console.log(error)
                )
        )

        // Run the check upon initial mount, so the user sees the information without needing to click.
        that.checkRequests.next()
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <BaseActionContainer
                title="Check connection to remote repository"
                description={<span>Diagnose problems cloning or updating from the remote repository.</span>}
                action={
                    <button
                        type="button"
                        className="btn btn-primary"
                        disabled={that.state.loading}
                        onClick={that.checkMirrorRepositoryConnection}
                    >
                        Check connection
                    </button>
                }
                details={
                    <>
                        {that.state.errorDescription && (
                            <ErrorAlert className="action-container__alert" error={that.state.errorDescription} />
                        )}
                        {that.state.loading && (
                            <div className="alert alert-primary action-container__alert">
                                <LoadingSpinner className="icon-inline" /> Checking connection...
                            </div>
                        )}
                        {that.state.result &&
                            (that.state.result.error === null ? (
                                <div className="alert alert-success action-container__alert">
                                    <CheckIcon className="icon-inline" /> The remote repository is reachable.
                                </div>
                            ) : (
                                <div className="alert alert-danger action-container__alert">
                                    <p>The remote repository is unreachable. Logs follow.</p>
                                    <div>
                                        <pre className="check-mirror-repository-connection-action-container__log">
                                            <code>{that.state.result.error}</code>
                                        </pre>
                                    </div>
                                </div>
                            ))}
                    </>
                }
            />
        )
    }

    private checkMirrorRepositoryConnection = (): void => that.checkRequests.next()
}

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
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

        that.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsMirror')

        that.subscriptions.add(
            that.repoUpdates.pipe(switchMap(() => fetchRepository(that.props.repo.name))).subscribe(
                repo => that.setState({ repo }),
                err => that.setState({ error: err.message })
            )
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-settings-mirror-page">
                <PageTitle title="Mirror settings" />
                <h2>Mirroring and cloning</h2>
                {that.state.loading && <LoadingSpinner className="icon-inline" />}
                {that.state.error && <ErrorAlert error={that.state.error} />}
                <div className="form-group">
                    <label>
                        Remote repository URL{' '}
                        <small className="text-info">
                            <LockIcon className="icon-inline" /> Only visible to site admins
                        </small>
                    </label>
                    <input
                        className="form-control"
                        value={that.props.repo.mirrorInfo.remoteURL || '(unknown)'}
                        readOnly={true}
                    />
                    {that.state.repo.viewerCanAdminister && (
                        <small className="form-text text-muted">
                            Configure repository mirroring in{' '}
                            <Link to="/site-admin/external-services">external services</Link>.
                        </small>
                    )}
                </div>
                <UpdateMirrorRepositoryActionContainer
                    repo={that.state.repo}
                    onDidUpdateRepository={that.onDidUpdateRepository}
                    disabled={typeof that.state.reachable === 'boolean' && !that.state.reachable}
                    disabledReason={
                        typeof that.state.reachable === 'boolean' && !that.state.reachable ? 'Not reachable' : undefined
                    }
                />
                <CheckMirrorRepositoryConnectionActionContainer
                    repo={that.state.repo}
                    onDidUpdateReachability={that.onDidUpdateReachability}
                />
                {typeof that.state.reachable === 'boolean' && !that.state.reachable && (
                    <div className="alert alert-info repo-settings-mirror-page__troubleshooting">
                        Problems cloning or updating that repository?
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
                                <Link to="/help/admin/repo/auth#ssh-authentication-config-keys-known-hosts">
                                    SSH repository authentication documentation
                                </Link>{' '}
                                for how to provide an SSH <code>known_hosts</code> file with the remote host's SSH host
                                key.
                            </li>
                            <li className="repo-settings-mirror-page__step">
                                Consult <Link to="/help/admin/repo/add">Sourcegraph repositories documentation</Link>{' '}
                                for resolving other authentication issues (such as HTTPS certificates and SSH keys).
                            </li>
                            <li className="repo-settings-mirror-page__step">
                                <FeedbackText headerText="Questions?" />
                            </li>
                        </ul>
                    </div>
                )}
            </div>
        )
    }

    private onDidUpdateRepository = (): void => {
        this.repoUpdates.next()
        this.props.onDidUpdateRepository({})
    }

    private onDidUpdateReachability = (reachable: boolean | undefined): void => this.setState({ reachable })
}
