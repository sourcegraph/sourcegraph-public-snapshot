import React, { type FC, useEffect, useState } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiLock } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useMutation, useQuery } from '@sourcegraph/http-client'
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
    ErrorAlert,
    CollapseHeader,
    Collapse,
    CollapsePanel,
    Label,
} from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import type {
    CheckMirrorRepositoryConnectionResult,
    CheckMirrorRepositoryConnectionVariables,
    RecloneRepositoryResult,
    RecloneRepositoryVariables,
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    UpdateMirrorRepositoryResult,
    UpdateMirrorRepositoryVariables,
} from '../../graphql-operations'
import {
    CHECK_MIRROR_REPOSITORY_CONNECTION,
    RECLONE_REPOSITORY_MUTATION,
    UPDATE_MIRROR_REPOSITORY,
} from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { DirectImportRepoAlert } from '../DirectImportRepoAlert'

import { FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'
import { RepoSettingsOptions } from './RepoSettingsOptions'

import styles from './RepoSettingsMirrorPage.module.scss'

interface UpdateMirrorRepositoryActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateRepository: () => Promise<void>
    disabled: boolean
    disabledReason: string | undefined
}

const UpdateMirrorRepositoryActionContainer: FC<UpdateMirrorRepositoryActionContainerProps> = props => {
    const [updateRepo] = useMutation<UpdateMirrorRepositoryResult, UpdateMirrorRepositoryVariables>(
        UPDATE_MIRROR_REPOSITORY,
        { variables: { repository: props.repo.id } }
    )

    const run = async (): Promise<void> => {
        await updateRepo()
        await props.onDidUpdateRepository()
    }

    let title: React.ReactNode
    let description: React.ReactNode
    let buttonLabel: React.ReactNode
    let buttonDisabled = false
    let info: React.ReactNode
    if (props.repo.mirrorInfo.cloneInProgress) {
        title = 'Cloning in progress...'
        description = props.repo.mirrorInfo.cloneProgress ? (
            <div className="overflow-auto">
                <Code>{props.repo.mirrorInfo.cloneProgress}</Code>
            </div>
        ) : (
            'This repository is currently being cloned from its remote repository.'
        )
        buttonLabel = (
            <span>
                <LoadingSpinner /> Cloning...
            </span>
        )
        buttonDisabled = true
        info = <DirectImportRepoAlert className={classNames(styles.alert, 'mb-0')} />
    } else if (props.repo.mirrorInfo.cloned) {
        const updateSchedule = props.repo.mirrorInfo.updateSchedule
        title = (
            <>
                <div>
                    Last refreshed:{' '}
                    {props.repo.mirrorInfo.updatedAt ? <Timestamp date={props.repo.mirrorInfo.updatedAt} /> : 'unknown'}{' '}
                </div>
            </>
        )
        info = (
            <>
                {updateSchedule && (
                    <div>
                        Next scheduled update <Timestamp date={updateSchedule.due} /> (position{' '}
                        {updateSchedule.index + 1} out of {updateSchedule.total} in the schedule)
                    </div>
                )}
                {props.repo.mirrorInfo.updateQueue && !props.repo.mirrorInfo.updateQueue.updating && (
                    <div>
                        Queued for update (position {props.repo.mirrorInfo.updateQueue.index + 1} out of{' '}
                        {props.repo.mirrorInfo.updateQueue.total} in the queue)
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
            titleAs="h3"
            description={description}
            buttonLabel={buttonLabel}
            buttonDisabled={buttonDisabled || props.disabled}
            buttonSubtitle={props.disabledReason}
            flashText="Added to queue"
            info={info}
            run={run}
        />
    )
}

interface CheckMirrorRepositoryConnectionActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateReachability: (reachable: boolean) => void
}

const CheckMirrorRepositoryConnectionActionContainer: FC<
    CheckMirrorRepositoryConnectionActionContainerProps
> = props => {
    const [checkConnection, { data, loading, error }] = useMutation<
        CheckMirrorRepositoryConnectionResult,
        CheckMirrorRepositoryConnectionVariables
    >(CHECK_MIRROR_REPOSITORY_CONNECTION, {
        variables: { repository: props.repo.id },
        onCompleted: result => {
            props.onDidUpdateReachability(result.checkMirrorRepositoryConnection.error === null)
        },
        onError: () => {
            props.onDidUpdateReachability(false)
        },
    })

    useEffect(() => {
        checkConnection().catch(() => {})
    }, [checkConnection])

    return (
        <BaseActionContainer
            title="Check connection to remote repository"
            titleAs="h3"
            description={<span>Diagnose problems cloning or updating from the remote repository.</span>}
            action={
                <Button
                    disabled={loading}
                    onClick={() => {
                        checkConnection().catch(() => {})
                    }}
                    variant="primary"
                >
                    Check connection
                </Button>
            }
            details={
                <>
                    {error && <ErrorAlert className={styles.alert} error={error} />}
                    {loading && (
                        <Alert className={classNames('mb-0', styles.alert)} variant="primary">
                            <LoadingSpinner /> Checking connection...
                        </Alert>
                    )}
                    {data &&
                        !loading &&
                        (data.checkMirrorRepositoryConnection.error === null ? (
                            <Alert className={classNames('mb-0', styles.alert)} variant="success">
                                The remote repository is reachable.
                            </Alert>
                        ) : (
                            <Alert className={classNames('mb-0', styles.alert)} variant="danger">
                                <Text>The remote repository is unreachable. Logs follow.</Text>
                                <div>
                                    <pre className={styles.log}>
                                        <Code>{data.checkMirrorRepositoryConnection.error}</Code>
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

// Add interface for props then create component
interface CorruptionLogProps {
    repo: SettingsAreaRepositoryFields
}

const CorruptionLogsContainer: FC<CorruptionLogProps> = props => {
    const health = props.repo.mirrorInfo.isCorrupted ? (
        <>
            <Alert className={classNames('mb-0', styles.alert)} variant="danger">
                The repository is corrupt, check the log entries below for more info and consider recloning.
            </Alert>
            <br />
        </>
    ) : null

    const logEvents: JSX.Element[] = props.repo.mirrorInfo.corruptionLogs.map(log => (
        <li key={`${props.repo.name}#${log.timestamp}`} className="list-group-item px-2 py-1">
            <div className="d-flex flex-column align-items-center justify-content-between">
                <Text className={classNames('overflow-auto', 'text-monospace', styles.log)}>{log.reason}</Text>
                <small className="text-muted mb-0">
                    <Timestamp date={log.timestamp} />
                </small>
            </div>
        </li>
    ))

    const [isOpened, setIsOpened] = useState(false)
    const hasLogs = logEvents.length !== 0

    return (
        <BaseActionContainer
            title="Repository corruption"
            titleAs="h3"
            description={<span>Recent corruption events that have been detected on this repository.</span>}
            className="mb-0"
            details={
                <div className="flex-1">
                    {health}
                    {!hasLogs && <Text className="mt-3 text-muted text-center mb-0">No corruption history</Text>}
                    {hasLogs && (
                        <Collapse isOpen={isOpened} onOpenChange={setIsOpened}>
                            <CollapseHeader
                                as={Button}
                                outline={true}
                                focusLocked={true}
                                variant="secondary"
                                className="w-100 my-2"
                                disabled={!hasLogs}
                            >
                                Show corruption history
                                <Icon
                                    aria-hidden={true}
                                    svgPath={isOpened ? mdiChevronUp : mdiChevronDown}
                                    className="mr-1"
                                />
                            </CollapseHeader>
                            <CollapsePanel>
                                <ul className="list-group">{logEvents}</ul>
                            </CollapsePanel>
                        </Collapse>
                    )}
                </div>
            }
        />
    )
}

interface RepoSettingsMirrorPageProps {
    repo: SettingsAreaRepositoryFields
    disablePolling?: boolean
}

/**
 * The repository settings mirror page.
 */
export const RepoSettingsMirrorPage: FC<RepoSettingsMirrorPageProps> = ({
    repo: initialRepo,
    disablePolling = false,
}) => {
    eventLogger.logPageView('RepoSettingsMirror')
    const [reachable, setReachable] = useState<boolean>()
    const [recloneRepository] = useMutation<RecloneRepositoryResult, RecloneRepositoryVariables>(
        RECLONE_REPOSITORY_MUTATION,
        {
            variables: { repo: initialRepo.id },
        }
    )

    const { data, error, refetch } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        {
            variables: { name: initialRepo.name },
            pollInterval: disablePolling ? undefined : 3000,
        }
    )

    const repo = data?.repository ? data.repository : initialRepo

    const onDidUpdateReachability = (reachable: boolean | undefined): void => setReachable(reachable)

    return (
        <>
            <PageTitle title="Mirror settings" />
            <PageHeader path={[{ text: 'Mirroring and cloning' }]} headingElement="h2" className="mb-3" />
            <RepoSettingsOptions repo={repo} />
            <Container className="repo-settings-mirror-page">
                {error && <ErrorAlert error={error} />}

                <div className="form-group">
                    <Label>
                        {' '}
                        Remote repository URL{' '}
                        <small className="text-muted">
                            <Icon aria-hidden={true} svgPath={mdiLock} className="text-warning" /> Only visible to site
                            admins
                        </small>
                    </Label>
                    <Input value={repo.mirrorInfo.remoteURL || '(unknown)'} readOnly={true} className="mb-0" />
                    {repo.viewerCanAdminister && (
                        <small className="form-text text-muted">
                            Configure repository mirroring in{' '}
                            <Link to="/site-admin/external-services">code host connections</Link>.
                        </small>
                    )}
                </div>
                {repo.mirrorInfo.lastError && (
                    <Alert variant="warning">
                        {/* TODO: This should not be a list item, but it was before this was refactored. */}
                        <li className="d-flex w-100">Error updating repo:</li>
                        <li className="d-flex w-100">{repo.mirrorInfo.lastError}</li>
                    </Alert>
                )}
                <UpdateMirrorRepositoryActionContainer
                    repo={repo}
                    onDidUpdateRepository={async () => {
                        await refetch()
                    }}
                    disabled={typeof reachable === 'boolean' && !reachable}
                    disabledReason={typeof reachable === 'boolean' && !reachable ? 'Not reachable' : undefined}
                />
                <ActionContainer
                    title="Reclone repository"
                    titleAs="h3"
                    description={
                        <div>
                            This will delete the repository from disk and reclone it.
                            <div className="mt-2">
                                <span className="font-weight-bold text-danger">WARNING</span>: This can take a long
                                time, depending on how large the repository is. The repository will be unsearchable
                                while the reclone is in progress.
                            </div>
                        </div>
                    }
                    buttonVariant="danger"
                    buttonLabel={
                        repo.mirrorInfo.cloneInProgress ? (
                            <span>
                                <LoadingSpinner /> Cloning...
                            </span>
                        ) : (
                            'Reclone'
                        )
                    }
                    buttonDisabled={repo.mirrorInfo.cloneInProgress}
                    flashText="Recloning repo"
                    run={async () => {
                        await recloneRepository()
                    }}
                />
                <CheckMirrorRepositoryConnectionActionContainer
                    repo={repo}
                    onDidUpdateReachability={onDidUpdateReachability}
                />
                {reachable === false && (
                    <Alert variant="info">
                        Problems cloning or updating this repository?
                        <ul className={styles.steps}>
                            <li className={styles.step}>
                                Inspect the <strong>Check connection</strong> error log output to see why the remote
                                repository is not reachable.
                            </li>
                            <li className={styles.step}>
                                <Code weight="bold">No ECDSA host key is known ... Host key verification failed?</Code>{' '}
                                See{' '}
                                <Link to="/help/admin/repo/auth#ssh-authentication-config-keys-known-hosts">
                                    SSH repository authentication documentation
                                </Link>{' '}
                                for how to provide an SSH <Code>known_hosts</Code> file with the remote host's SSH host
                                key.
                            </li>
                            <li className={styles.step}>
                                Consult <Link to="/help/admin/repo/add">Sourcegraph repositories documentation</Link>{' '}
                                for resolving other authentication issues (such as HTTPS certificates and SSH keys).
                            </li>
                            <li className={styles.step}>
                                <FeedbackText headerText="Questions?" />
                            </li>
                        </ul>
                    </Alert>
                )}
                <CorruptionLogsContainer repo={repo} />
            </Container>
        </>
    )
}
