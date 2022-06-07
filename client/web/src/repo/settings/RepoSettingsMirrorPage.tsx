import * as React from 'react'

import classNames from 'classnames'
import * as H from 'history'
import LockIcon from 'mdi-react/LockIcon'
import { RouteComponentProps } from 'react-router'
import { interval, Subject, Subscription } from 'rxjs'
import { catchError, switchMap, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError } from '@sourcegraph/common'
import * as GQL from '@sourcegraph/shared/src/schema'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    FeedbackText,
    Button,
    Link,
    Alert,
    Icon,
    Input,
    Text,
    Code,
} from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { checkMirrorRepositoryConnection, updateMirrorRepository } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { DirectImportRepoAlert } from '../DirectImportRepoAlert'

import { fetchSettingsAreaRepository } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'

import styles from './RepoSettingsMirrorPage.module.scss'

interface UpdateMirrorRepositoryActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateRepository: () => void
    disabled: boolean
    disabledReason: string | undefined
    history: H.History
}

class UpdateMirrorRepositoryActionContainer extends React.PureComponent<UpdateMirrorRepositoryActionContainerProps> {
    private componentUpdates = new Subject<UpdateMirrorRepositoryActionContainerProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            interval(3000).subscribe(() => {
                this.props.onDidUpdateRepository()
            })
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let title: React.ReactFragment
        let description: React.ReactFragment
        let buttonLabel: React.ReactFragment
        let buttonDisabled = false
        let info: React.ReactNode
        if (this.props.repo.mirrorInfo.cloneInProgress) {
            title = 'Cloning in progress...'
            description =
                <Code>{this.props.repo.mirrorInfo.cloneProgress}</Code> ||
                'This repository is currently being cloned from its remote repository.'
            buttonLabel = (
                <span>
                    <LoadingSpinner /> Cloning...
                </span>
            )
            buttonDisabled = true
            info = <DirectImportRepoAlert className={styles.alert} />
        } else if (this.props.repo.mirrorInfo.cloned) {
            const updateSchedule = this.props.repo.mirrorInfo.updateSchedule
            title = (
                <>
                    <div>
                        Last refreshed:{' '}
                        {this.props.repo.mirrorInfo.updatedAt ? (
                            <Timestamp date={this.props.repo.mirrorInfo.updatedAt} />
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
                    {this.props.repo.mirrorInfo.updateQueue && !this.props.repo.mirrorInfo.updateQueue.updating && (
                        <div>
                            Queued for update (position {this.props.repo.mirrorInfo.updateQueue.index + 1} out of{' '}
                            {this.props.repo.mirrorInfo.updateQueue.total} in the queue)
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
                buttonDisabled={buttonDisabled || this.props.disabled}
                buttonSubtitle={this.props.disabledReason}
                flashText="Added to queue"
                info={info}
                run={this.updateMirrorRepository}
                history={this.props.history}
            />
        )
    }

    private updateMirrorRepository = async (): Promise<void> => {
        await updateMirrorRepository({ repository: this.props.repo.id }).toPromise()
        this.props.onDidUpdateRepository()
    }
}

interface CheckMirrorRepositoryConnectionActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateReachability: (reachable: boolean | undefined) => void
    history: H.History
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
                                this.setState({
                                    errorDescription: asError(error).message,
                                    result: undefined,
                                    loading: false,
                                })
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
                    <Button
                        disabled={this.state.loading}
                        onClick={this.checkMirrorRepositoryConnection}
                        variant="primary"
                    >
                        Check connection
                    </Button>
                }
                details={
                    <>
                        {this.state.errorDescription && (
                            <ErrorAlert className={styles.alert} error={this.state.errorDescription} />
                        )}
                        {this.state.loading && (
                            <Alert className={classNames('mb-0', styles.alert)} variant="primary">
                                <LoadingSpinner /> Checking connection...
                            </Alert>
                        )}
                        {this.state.result &&
                            (this.state.result.error === null ? (
                                <Alert className={classNames('mb-0', styles.alert)} variant="success">
                                    The remote repository is reachable.
                                </Alert>
                            ) : (
                                <Alert className={classNames('mb-0', styles.alert)} variant="danger">
                                    <Text>The remote repository is unreachable. Logs follow.</Text>
                                    <div>
                                        <pre className={styles.log}>
                                            <Code>{this.state.result.error}</Code>
                                        </pre>
                                    </div>
                                </Alert>
                            ))}
                    </>
                }
                className="mb-0"
            />
        )
    }

    private checkMirrorRepositoryConnection = (): void => this.checkRequests.next()
}

interface RepoSettingsMirrorPageProps extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
    history: H.History
}

interface RepoSettingsMirrorPageState {
    /**
     * The repository object, refreshed after we make changes that modify it.
     */
    repo: SettingsAreaRepositoryFields

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
export class RepoSettingsMirrorPage extends React.PureComponent<
    RepoSettingsMirrorPageProps,
    RepoSettingsMirrorPageState
> {
    private repoUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: RepoSettingsMirrorPageProps) {
        super(props)

        this.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsMirror')

        this.subscriptions.add(
            this.repoUpdates.pipe(switchMap(() => fetchSettingsAreaRepository(this.props.repo.name))).subscribe(
                repo => this.setState({ repo }),
                error => this.setState({ error: asError(error).message })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <>
                <PageTitle title="Mirror settings" />
                <PageHeader path={[{ text: 'Mirroring and cloning' }]} headingElement="h2" className="mb-3" />
                <Container className="repo-settings-mirror-page">
                    {this.state.loading && <LoadingSpinner />}
                    {this.state.error && <ErrorAlert error={this.state.error} />}
                    <div className="form-group">
                        <Input
                            value={this.props.repo.mirrorInfo.remoteURL || '(unknown)'}
                            readOnly={true}
                            className="mb-0"
                            label={
                                <>
                                    {' '}
                                    Remote repository URL{' '}
                                    <small className="text-info">
                                        <Icon role="img" as={LockIcon} aria-hidden={true} /> Only visible to site admins
                                    </small>
                                </>
                            }
                        />
                        {this.state.repo.viewerCanAdminister && (
                            <small className="form-text text-muted">
                                Configure repository mirroring in{' '}
                                <Link to="/site-admin/external-services">external services</Link>.
                            </small>
                        )}
                    </div>
                    <UpdateMirrorRepositoryActionContainer
                        repo={this.state.repo}
                        onDidUpdateRepository={this.onDidUpdateRepository}
                        disabled={typeof this.state.reachable === 'boolean' && !this.state.reachable}
                        disabledReason={
                            typeof this.state.reachable === 'boolean' && !this.state.reachable
                                ? 'Not reachable'
                                : undefined
                        }
                        history={this.props.history}
                    />
                    <CheckMirrorRepositoryConnectionActionContainer
                        repo={this.state.repo}
                        onDidUpdateReachability={this.onDidUpdateReachability}
                        history={this.props.history}
                    />
                    {typeof this.state.reachable === 'boolean' && !this.state.reachable && (
                        <Alert variant="info">
                            Problems cloning or updating this repository?
                            <ul className={styles.steps}>
                                <li className={styles.step}>
                                    Inspect the <strong>Check connection</strong> error log output to see why the remote
                                    repository is not reachable.
                                </li>
                                <li className={styles.step}>
                                    <Code weight="bold">
                                        No ECDSA host key is known ... Host key verification failed?
                                    </Code>{' '}
                                    See{' '}
                                    <Link to="/help/admin/repo/auth#ssh-authentication-config-keys-known-hosts">
                                        SSH repository authentication documentation
                                    </Link>{' '}
                                    for how to provide an SSH <Code>known_hosts</Code> file with the remote host's SSH
                                    host key.
                                </li>
                                <li className={styles.step}>
                                    Consult{' '}
                                    <Link to="/help/admin/repo/add">Sourcegraph repositories documentation</Link> for
                                    resolving other authentication issues (such as HTTPS certificates and SSH keys).
                                </li>
                                <li className={styles.step}>
                                    <FeedbackText headerText="Questions?" />
                                </li>
                            </ul>
                        </Alert>
                    )}
                </Container>
            </>
        )
    }

    private onDidUpdateRepository = (): void => {
        this.repoUpdates.next()
    }

    private onDidUpdateReachability = (reachable: boolean | undefined): void => this.setState({ reachable })
}
