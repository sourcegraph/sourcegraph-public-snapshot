import React, { useEffect, useState } from 'react'

import { mdiLock } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { RouteComponentProps } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
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

import { TerminalLine } from '../../auth/Terminal'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import {
    RecloneRepositoryResult,
    RecloneRepositoryVariables,
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
} from '../../graphql-operations'
import {
    checkMirrorRepositoryConnection,
    RECLONE_REPOSITORY_MUTATION,
    updateMirrorRepository,
} from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'
import { DirectImportRepoAlert } from '../DirectImportRepoAlert'

import { FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'
import { ActionContainer, BaseActionContainer } from './components/ActionContainer'

import styles from './RepoSettingsMirrorPage.module.scss'

interface UpdateMirrorRepositoryActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateRepository: () => void
    disabled: boolean
    disabledReason: string | undefined
    history: H.History
}

const UpdateMirrorRepositoryActionContainer: React.FunctionComponent<
    UpdateMirrorRepositoryActionContainerProps
> = props => {
    const thisUpdateMirrorRepository = async (): Promise<void> => {
        await updateMirrorRepository({ repository: props.repo.id }).toPromise()
        props.onDidUpdateRepository()
    }

    let title: React.ReactNode
    let description: React.ReactNode
    let buttonLabel: React.ReactNode
    let buttonDisabled = false
    let info: React.ReactNode
    if (props.repo.mirrorInfo.cloneInProgress) {
        title = 'Cloning in progress...'
        description =
            <Code>{props.repo.mirrorInfo.cloneProgress}</Code> ||
            'This repository is currently being cloned from its remote repository.'
        buttonLabel = (
            <span>
                <LoadingSpinner /> Cloning...
            </span>
        )
        buttonDisabled = true
        info = <DirectImportRepoAlert className={styles.alert} />
    } else if (props.repo.mirrorInfo.cloned) {
        const updateSchedule = props.repo.mirrorInfo.updateSchedule
        title = (
            <>
                <div>
                    Last refreshed:{' '}
                    {props.repo.mirrorInfo.updatedAt ? <Timestamp date={props.repo.mirrorInfo.updatedAt} /> : 'unknown'}{' '}
                </div>
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
            description={<div>{description}</div>}
            buttonLabel={buttonLabel}
            buttonDisabled={buttonDisabled || props.disabled}
            buttonSubtitle={props.disabledReason}
            flashText="Added to queue"
            info={info}
            run={thisUpdateMirrorRepository}
            history={props.history}
        />
    )
}

interface CheckMirrorRepositoryConnectionActionContainerProps {
    repo: SettingsAreaRepositoryFields
    onDidUpdateReachability: (reachable: boolean | undefined) => void
    history: H.History
}

const CheckMirrorRepositoryConnectionActionContainer: React.FunctionComponent<
    CheckMirrorRepositoryConnectionActionContainerProps
> = props => {
    const [loading, setLoading] = useState<boolean>(false)
    const [errorDescription, setErrorDescription] = useState<string>()
    const [result, setResult] = useState<GQL.ICheckMirrorRepositoryConnectionResult>()

    const thisCheckMirrorRepositoryConnection = (): void => {
        checkMirrorRepositoryConnection({ repository: props.repo.id })
            .toPromise()
            .then(result => {
                setResult(result)
                setLoading(false)
                props.onDidUpdateReachability(result.error === null)
            })
            .catch(error => {
                setLoading(false)
                setErrorDescription(asError(error).message)
                setResult(undefined)
                props.onDidUpdateReachability(false)
                return []
            })
    }
    useEffect(() => {
        setLoading(true)
        setErrorDescription(undefined)
        setResult(undefined)
        props.onDidUpdateReachability(undefined)
        thisCheckMirrorRepositoryConnection()
    }, [])

    return (
        <BaseActionContainer
            title="Check connection to remote repository"
            description={<span>Diagnose problems cloning or updating from the remote repository.</span>}
            action={
                <Button disabled={loading} onClick={thisCheckMirrorRepositoryConnection} variant="primary">
                    Check connection
                </Button>
            }
            details={
                <>
                    {errorDescription && <ErrorAlert className={styles.alert} error={errorDescription} />}
                    {loading && (
                        <Alert className={classNames('mb-0', styles.alert)} variant="primary">
                            <LoadingSpinner /> Checking connection...
                        </Alert>
                    )}
                    {result &&
                        (result.error === null ? (
                            <Alert className={classNames('mb-0', styles.alert)} variant="success">
                                The remote repository is reachable.
                            </Alert>
                        ) : (
                            <Alert className={classNames('mb-0', styles.alert)} variant="danger">
                                <Text>The remote repository is unreachable. Logs follow.</Text>
                                <div>
                                    <pre className={styles.log}>
                                        <Code>{result.error}</Code>
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

interface RepoSettingsMirrorPageProps extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
    history: H.History
}

/**
 * The repository settings mirror page.
 */
export const RepoSettingsMirrorPage: React.FunctionComponent<
    React.PropsWithChildren<RepoSettingsMirrorPageProps>
> = props => {
    const [repo, setRepo] = useState<SettingsAreaRepositoryFields>(props.repo)
    const [reachable, setReachable] = useState<boolean>()
    const [error, setError] = useState<string>()
    const [recloneRepository] = useMutation<RecloneRepositoryResult, RecloneRepositoryVariables>(
        RECLONE_REPOSITORY_MUTATION,
        {
            variables: { repo: repo.id },
        }
    )

    const fetchRepo = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        {
            variables: { name: repo.name },
        }
    )

    const updateRepo = async (): Promise<void> => {
        const { data, error } = await fetchRepo.refetch()

        if (data?.repository) {
            setRepo(data.repository)
        }

        if (error) {
            setError(error.message)
        }
    }

    useEffect(() => {
        eventLogger.logPageView('RepoSettingsMirror')
        updateRepo()

        setInterval(() => {
            updateRepo()
        }, 3000)
    }, [])

    const onDidUpdateRepository = (): void => {
        updateRepo()
    }

    const onDidUpdateReachability = (reachable: boolean | undefined): void => setReachable(reachable)

    return (
        <>
            <PageTitle title="Mirror settings" />
            <PageHeader path={[{ text: 'Mirroring and cloning' }]} headingElement="h2" className="mb-3" />
            <Container className="repo-settings-mirror-page">
                {error && <ErrorAlert error={error} />}

                <div className="form-group">
                    <Input
                        value={repo.mirrorInfo.remoteURL || '(unknown)'}
                        readOnly={true}
                        className="mb-0"
                        label={
                            <>
                                {' '}
                                Remote repository URL{' '}
                                <small className="text-info">
                                    <Icon aria-hidden={true} svgPath={mdiLock} /> Only visible to site admins
                                </small>
                            </>
                        }
                    />
                    {repo.viewerCanAdminister && (
                        <small className="form-text text-muted">
                            Configure repository mirroring in{' '}
                            <Link to="/site-admin/external-services">external services</Link>.
                        </small>
                    )}
                </div>
                {repo.mirrorInfo.lastError && (
                    <Alert variant="warning">
                        <TerminalLine>Error updating repo:</TerminalLine>
                        <TerminalLine>{repo.mirrorInfo.lastError}</TerminalLine>
                    </Alert>
                )}
                <UpdateMirrorRepositoryActionContainer
                    repo={repo}
                    onDidUpdateRepository={onDidUpdateRepository}
                    disabled={typeof reachable === 'boolean' && !reachable}
                    disabledReason={typeof reachable === 'boolean' && !reachable ? 'Not reachable' : undefined}
                    history={props.history}
                />
                <ActionContainer
                    title="Reclone repository"
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
                    history={props.history}
                />
                <CheckMirrorRepositoryConnectionActionContainer
                    repo={repo}
                    onDidUpdateReachability={onDidUpdateReachability}
                    history={props.history}
                />
                {typeof reachable === 'boolean' && !reachable && (
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
            </Container>
        </>
    )
}
