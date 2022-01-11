import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { Redirect, Route, RouteComponentProps, Switch, useHistory, useLocation } from 'react-router'
import { NavLink as RouterLink } from 'react-router-dom'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, LoadingSpinner, PageHeader, FeedbackBadge } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { Duration } from '../../../components/time/Duration'
import { Timestamp } from '../../../components/time/Timestamp'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
    Scalars,
} from '../../../graphql-operations'
import { BatchSpec } from '../BatchSpec'
import { NewBatchChangePreviewPage } from '../preview/BatchChangePreviewPage'

import { cancelBatchSpecExecution, FETCH_BATCH_SPEC_EXECUTION, retryBatchSpecExecution } from './backend'
import styles from './BatchSpecExecutionDetailsPage.module.scss'
import { BatchSpecStateBadge } from './BatchSpecStateBadge'
import { WorkspaceDetails } from './WorkspaceDetails'
import { WorkspacesList } from './WorkspacesList'

export interface BatchSpecExecutionDetailsPageProps extends ThemeProps, TelemetryProps, RouteComponentProps<{}> {
    batchSpecID: Scalars['ID']
    authenticatedUser: AuthenticatedUser
}

export const BatchSpecExecutionDetailsPage: React.FunctionComponent<BatchSpecExecutionDetailsPageProps> = ({
    batchSpecID,
    isLightTheme,
    authenticatedUser,
    telemetryService,
    match,
}) => {
    const { data, error, loading } = useQuery<BatchSpecExecutionByIDResult, BatchSpecExecutionByIDVariables>(
        FETCH_BATCH_SPEC_EXECUTION,
        {
            variables: { id: batchSpecID },
            fetchPolicy: 'cache-and-network',
            pollInterval: 2500,
            nextFetchPolicy: 'network-only',
        }
    )

    if (loading) {
        return (
            <div className="text-center">
                <LoadingSpinner className="mx-auto my-4" />
            </div>
        )
    }

    if (error) {
        return <HeroPage icon={AlertCircleIcon} title={String(error)} />
    }

    if (!data?.node || data.node.__typename !== 'BatchSpec') {
        return <HeroPage icon={AlertCircleIcon} title="Execution not found" />
    }

    const batchSpec = data.node

    return (
        <div className="d-flex flex-column p-4 w-100 h-100">
            <PageTitle title="Batch spec execution" />
            <PageHeader
                path={[
                    {
                        icon: BatchChangesIcon,
                        to: '/batch-changes',
                    },
                    {
                        to: `${batchSpec.namespace.url}/batch-changes`,
                        text: batchSpec.namespace.namespaceName,
                    },
                    // If a matching batch change already exists, link to it.
                    {
                        to: batchSpec.appliesToBatchChange?.url ?? undefined,
                        text: batchSpec.description.name,
                    },
                ]}
                annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                byline={
                    <>
                        Created <Timestamp date={batchSpec.createdAt} /> by{' '}
                        <LinkOrSpan to={batchSpec.creator?.url}>
                            {batchSpec.creator?.displayName || batchSpec.creator?.username || 'a deleted user'}
                        </LinkOrSpan>
                    </>
                }
                actions={<BatchSpecActions batchSpec={batchSpec} />}
                className="mb-3"
            />

            <TabBar url={match.url} batchSpec={batchSpec} />

            <Switch>
                <Route render={() => <Redirect to={`${match.url}/execution`} />} path={match.url} exact={true} />
                <Route
                    path={`${match.url}/edit`}
                    render={() => (
                        <EditPage
                            name={batchSpec.description.name}
                            content={batchSpec.originalInput}
                            isLightTheme={isLightTheme}
                        />
                    )}
                    exact={true}
                />
                <Route
                    path={`${match.url}/execution`}
                    render={() => <ExecutionPage batchSpec={batchSpec} isLightTheme={isLightTheme} />}
                    exact={true}
                />
                <Route
                    path={`${match.url}/preview`}
                    render={() => (
                        <PreviewPage
                            batchSpec={batchSpec}
                            authenticatedUser={authenticatedUser}
                            batchSpecID={batchSpec.id}
                            isLightTheme={isLightTheme}
                            telemetryService={telemetryService}
                        />
                    )}
                    exact={true}
                />
                <Route component={NotFoundPage} key="hardcoded-key" />
            </Switch>
        </div>
    )
}

const TabBar: React.FunctionComponent<{ url: string; batchSpec: BatchSpecExecutionFields }> = ({ url, batchSpec }) => (
    <div className="mb-3">
        <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
            <li className="nav-item">
                <RouterLink to={`${url}/edit`} role="button" activeClassName="active" className="nav-link">
                    {/* TODO: Rename to edit once this IS an editor. */}
                    <span className="text-content" data-tab-content="1. Batch spec">
                        1. Batch spec
                    </span>
                </RouterLink>
            </li>
            <li className="nav-item">
                <RouterLink to={`${url}/execution`} role="button" activeClassName="active" className="nav-link">
                    <span className="text-content" data-tab-content="2. Execution">
                        2. Execution
                    </span>
                </RouterLink>
            </li>
            <li className="nav-item">
                {!batchSpec.applyURL && (
                    <span
                        aria-disabled="true"
                        className="nav-link text-muted"
                        data-tooltip="Wait for the execution to finish"
                    >
                        <span className="text-content" data-tab-content="3. Preview">
                            3. Preview
                        </span>
                    </span>
                )}
                {batchSpec.applyURL && (
                    <RouterLink to={`${url}/preview`} role="button" activeClassName="active" className="nav-link">
                        <span className="text-content" data-tab-content="3. Preview">
                            3. Preview
                        </span>
                    </RouterLink>
                )}
            </li>
        </ul>
    </div>
)

interface BatchSpecActionsProps {
    batchSpec: BatchSpecExecutionFields
}

const BatchSpecActions: React.FunctionComponent<BatchSpecActionsProps> = ({ batchSpec }) => {
    const location = useLocation()

    const [isCanceling, setIsCanceling] = useState<boolean | Error>(false)
    const cancelExecution = useCallback(async () => {
        try {
            // This reloads all the fields so apollo will rerender the parent component with new details too.
            // TODO: Actually use apollo here.
            await cancelBatchSpecExecution(batchSpec.id)
        } catch (error) {
            setIsCanceling(asError(error))
        }
    }, [batchSpec.id])

    const [isRetrying, setIsRetrying] = useState<boolean | Error>(false)
    const retryExecution = useCallback(async () => {
        try {
            // This reloads all the fields so apollo will rerender the parent component with new details too.
            // TODO: Actually use apollo here.
            await retryBatchSpecExecution(batchSpec.id)
        } catch (error) {
            setIsRetrying(asError(error))
        }
    }, [batchSpec.id])

    return (
        <div className="d-flex">
            <span className="align-self-center mr-2">
                <BatchSpecStateBadge state={batchSpec.state} />
            </span>
            {batchSpec.startedAt && (
                <div className="mx-2 text-center text-muted">
                    <h3>
                        <Duration start={batchSpec.startedAt} end={batchSpec.finishedAt ?? undefined} />
                    </h3>
                    Total time
                </div>
            )}
            {batchSpec.workspaceResolution?.workspaces.stats && (
                <>
                    <WorkspaceStat
                        stat={batchSpec.workspaceResolution.workspaces.stats.errored}
                        label="Errors"
                        iconClassName="text-danger"
                    />
                    <WorkspaceStat
                        stat={batchSpec.workspaceResolution.workspaces.stats.completed}
                        label="Complete"
                        iconClassName="text-success"
                    />
                    <WorkspaceStat stat={batchSpec.workspaceResolution.workspaces.stats.processing} label="Working" />
                    <WorkspaceStat stat={batchSpec.workspaceResolution.workspaces.stats.queued} label="Queued" />
                    <WorkspaceStat stat={batchSpec.workspaceResolution.workspaces.stats.ignored} label="Ignored" />
                </>
            )}
            <span>
                <div className="btn-group-vertical ml-2">
                    {(batchSpec.state === BatchSpecState.QUEUED || batchSpec.state === BatchSpecState.PROCESSING) && (
                        <Button
                            onClick={cancelExecution}
                            disabled={isCanceling === true}
                            outline={true}
                            variant="secondary"
                        >
                            {isCanceling !== true && <>Cancel</>}
                            {isCanceling === true && (
                                <>
                                    <LoadingSpinner /> Canceling
                                </>
                            )}
                        </Button>
                    )}
                    {!location.pathname.endsWith('preview') &&
                        batchSpec.applyURL &&
                        batchSpec.state === BatchSpecState.COMPLETED && (
                            <Link to="preview" className="btn btn-primary">
                                Preview
                            </Link>
                        )}
                    {batchSpec.viewerCanRetry && batchSpec.state !== BatchSpecState.COMPLETED && (
                        // TODO: Add a second button to allow retrying an entire batch spec,
                        // including completed jobs.
                        <Button
                            onClick={retryExecution}
                            disabled={isRetrying === true}
                            data-tooltip={isRetrying !== true ? 'Retry all failed workspaces' : undefined}
                            outline={true}
                            variant="secondary"
                        >
                            {isRetrying !== true && <>Retry</>}
                            {isRetrying === true && (
                                <>
                                    <LoadingSpinner className="icon-inline" /> Retrying
                                </>
                            )}
                        </Button>
                    )}
                    {!location.pathname.endsWith('preview') &&
                        batchSpec.applyURL &&
                        batchSpec.state === BatchSpecState.FAILED && (
                            <Link
                                className="btn btn-outline-warning"
                                to="preview"
                                data-tooltip="Execution didn't finish successfully in all workspaces. The batch spec might have less changeset specs than expected."
                            >
                                <AlertCircleIcon className="icon-inline mb-0 mr-2 text-warning" />
                                Preview
                            </Link>
                        )}
                </div>
                {isErrorLike(isCanceling) && <ErrorAlert error={isCanceling} />}
                {isErrorLike(isRetrying) && <ErrorAlert error={isRetrying} />}
            </span>
        </div>
    )
}

const WorkspaceStat: React.FunctionComponent<{ stat: number; label: string; iconClassName?: string }> = ({
    stat,
    label,
    iconClassName,
}) => (
    <div className="mx-2 text-center text-muted">
        <h3 className={iconClassName}>{stat}</h3>
        {label}
    </div>
)

interface EditPageProps extends ThemeProps {
    name: string
    content: string
}

const EditPage: React.FunctionComponent<EditPageProps> = ({ name, content, isLightTheme }) => (
    <div className={classNames(styles.layoutContainer, 'h-100')}>
        <BatchSpec name={name} originalInput={content} isLightTheme={isLightTheme} className={styles.batchSpec} />
    </div>
)

interface ExecutionPageProps extends ThemeProps {
    batchSpec: BatchSpecExecutionFields
}

const ExecutionPage: React.FunctionComponent<ExecutionPageProps> = ({ batchSpec, isLightTheme }) => {
    const history = useHistory()

    // Read the selected workspace from the URL params.
    const selectedWorkspace = useMemo(() => {
        const query = new URLSearchParams(history.location.search)
        return query.get('workspace')
    }, [history.location.search])

    return (
        <>
            {batchSpec.failureMessage && <ErrorAlert error={batchSpec.failureMessage} />}
            <div className={classNames(styles.layoutContainer, 'd-flex flex-1')}>
                <div className={classNames(styles.workspacesListContainer, 'd-flex flex-column')}>
                    <h3 className="mb-2">Workspaces</h3>
                    <div className={styles.workspacesList}>
                        <WorkspacesList batchSpecID={batchSpec.id} selectedNode={selectedWorkspace ?? undefined} />
                    </div>
                </div>
                <div className="d-flex flex-grow-1">
                    <div className="d-flex overflow-auto w-100">
                        <SelectedWorkspace workspace={selectedWorkspace} isLightTheme={isLightTheme} />
                    </div>
                </div>
            </div>
        </>
    )
}

interface PreviewPageProps extends TelemetryProps, ThemeProps {
    batchSpecID: Scalars['ID']
    batchSpec: BatchSpecExecutionFields
    authenticatedUser: AuthenticatedUser
}
const PreviewPage: React.FunctionComponent<PreviewPageProps> = ({
    authenticatedUser,
    telemetryService,
    isLightTheme,
    batchSpec,
    batchSpecID,
}) => {
    const history = useHistory()

    if (!batchSpec.applyURL) {
        return <Redirect to="./execution" />
    }

    return (
        <div className="mt-3">
            <NewBatchChangePreviewPage
                authenticatedUser={authenticatedUser}
                telemetryService={telemetryService}
                history={history}
                isLightTheme={isLightTheme}
                batchSpecID={batchSpecID}
                location={history.location}
            />
        </div>
    )
}

const SelectedWorkspace: React.FunctionComponent<{ workspace: Scalars['ID'] | null } & ThemeProps> = ({
    workspace,
    isLightTheme,
}) => {
    if (workspace === null) {
        return (
            <div className="card w-100">
                <div className="card-body">
                    <h3 className="text-center my-3">Select a workspace to view details.</h3>
                </div>
            </div>
        )
    }
    return (
        <div className="card w-100">
            <div className="card-body">
                <WorkspaceDetails id={workspace} isLightTheme={isLightTheme} />
            </div>
        </div>
    )
}

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />
