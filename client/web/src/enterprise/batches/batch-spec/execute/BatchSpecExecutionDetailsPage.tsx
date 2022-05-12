import React, { useCallback } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import CircleOffOutlineIcon from 'mdi-react/CircleOffOutlineIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimelineClockOutlineIcon from 'mdi-react/TimelineClockOutlineIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import { Redirect, Route, RouteComponentProps, Switch, useHistory, useLocation } from 'react-router'
import { NavLink as RouterLink } from 'react-router-dom'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { pluralize } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Button,
    LoadingSpinner,
    PageHeader,
    FeedbackBadge,
    ButtonGroup,
    Link,
    CardBody,
    Card,
    Icon,
    Panel,
    Typography,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { BatchChangesIcon } from '../../../../batches/icons'
import { HeroPage } from '../../../../components/HeroPage'
import { LoaderButton } from '../../../../components/LoaderButton'
import { PageTitle } from '../../../../components/PageTitle'
import { Duration } from '../../../../components/time/Duration'
import { Timestamp } from '../../../../components/time/Timestamp'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
    Scalars,
} from '../../../../graphql-operations'
import { BatchSpec } from '../../BatchSpec'
import { NewBatchChangePreviewPage } from '../../preview/BatchChangePreviewPage'

import { useCancelBatchSpecExecution, FETCH_BATCH_SPEC_EXECUTION, useRetryBatchSpecExecution } from './backend'
import { BatchSpecStateBadge } from './BatchSpecStateBadge'
import { WorkspaceDetails } from './workspaces/WorkspaceDetails'
import { Workspaces } from './workspaces/Workspaces'

import styles from './BatchSpecExecutionDetailsPage.module.scss'

export interface BatchSpecExecutionDetailsPageProps extends ThemeProps, TelemetryProps, RouteComponentProps<{}> {
    batchSpecID: Scalars['ID']
    authenticatedUser: AuthenticatedUser
}

export const BatchSpecExecutionDetailsPage: React.FunctionComponent<
    React.PropsWithChildren<BatchSpecExecutionDetailsPageProps>
> = ({ batchSpecID, isLightTheme, authenticatedUser, telemetryService, match }) => {
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
                        ariaLabel: 'Batch Changes Icon',
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
                actions={<BatchSpecActions batchSpec={batchSpec} executionURL={match.url} />}
                className="mb-3"
            />

            <TabBar url={match.url} batchSpec={batchSpec} />

            <Switch>
                <Route render={() => <Redirect to={`${match.url}/execution`} />} path={match.url} exact={true} />
                <Route
                    path={`${match.url}/spec`}
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
                    render={props => <ExecutionPage {...props} batchSpec={batchSpec} isLightTheme={isLightTheme} />}
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

const TabBar: React.FunctionComponent<
    React.PropsWithChildren<{ url: string; batchSpec: BatchSpecExecutionFields }>
> = ({ url, batchSpec }) => (
    <div className="mb-3">
        <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
            <li className="nav-item">
                <RouterLink to={`${url}/spec`} role="button" activeClassName="active" className="nav-link">
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
    executionURL: string
}

const BatchSpecActions: React.FunctionComponent<React.PropsWithChildren<BatchSpecActionsProps>> = ({
    batchSpec,
    executionURL,
}) => {
    const location = useLocation()

    const [cancelBatchSpecExecution, { loading: isCancelLoading, error: cancelError }] = useCancelBatchSpecExecution(
        batchSpec.id
    )

    const [retryBatchSpecExecution, { loading: isRetryLoading, error: retryError }] = useRetryBatchSpecExecution(
        batchSpec.id
    )

    const workspacesStats = batchSpec.workspaceResolution?.workspaces.stats

    return (
        <div className="d-flex">
            <span className="align-self-center mr-2" aria-label="Batch Spec status">
                <BatchSpecStateBadge state={batchSpec.state} />
            </span>
            {batchSpec.startedAt && (
                <div className={styles.workspacesStat}>
                    <ProgressClockIcon />
                    <Duration start={batchSpec.startedAt} end={batchSpec.finishedAt ?? undefined} />
                </div>
            )}
            {workspacesStats && (
                <>
                    <div className={styles.workspacesStat}>
                        <Icon as={AlertCircleIcon} className="text-danger" role="presentation" />
                        {`${workspacesStats.errored} ${pluralize('error', workspacesStats.errored)}`}
                    </div>
                    <div className={styles.workspacesStat}>
                        <Icon as={CheckBoldIcon} className="text-success" role="presentation" />
                        {`${workspacesStats.completed} complete`}
                    </div>
                    <div className={styles.workspacesStat}>
                        <Icon as={TimerSandIcon} role="presentation" />
                        {`${workspacesStats.processing} working`}
                    </div>
                    <div className={styles.workspacesStat}>
                        <Icon as={TimelineClockOutlineIcon} role="presentation" />
                        {`${workspacesStats.queued} queued`}
                    </div>
                    <div className={styles.workspacesStat}>
                        <Icon as={CircleOffOutlineIcon} role="presentation" />
                        {`${workspacesStats.ignored} ignored`}
                    </div>
                </>
            )}
            <span>
                <ButtonGroup direction="vertical" className="ml-2">
                    {(batchSpec.state === BatchSpecState.QUEUED || batchSpec.state === BatchSpecState.PROCESSING) && (
                        <LoaderButton
                            onClick={() => cancelBatchSpecExecution()}
                            disabled={isCancelLoading}
                            outline={true}
                            variant="danger"
                            loading={isCancelLoading}
                            alwaysShowLabel={true}
                            label="Cancel"
                        />
                    )}
                    {!location.pathname.endsWith('preview') &&
                        batchSpec.applyURL &&
                        batchSpec.state === BatchSpecState.COMPLETED && (
                            <Button to={`${executionURL}/preview`} variant="primary" as={Link}>
                                Preview
                            </Button>
                        )}
                    {batchSpec.viewerCanRetry && batchSpec.state !== BatchSpecState.COMPLETED && (
                        // TODO: Add a second button to allow retrying an entire batch spec,
                        // including completed jobs.
                        <LoaderButton
                            onClick={() => retryBatchSpecExecution()}
                            disabled={isRetryLoading}
                            data-tooltip={isRetryLoading ? undefined : 'Retry all failed workspaces'}
                            outline={true}
                            variant="secondary"
                            loading={isRetryLoading}
                            alwaysShowLabel={true}
                            label="Retry"
                        />
                    )}
                    {!location.pathname.endsWith('preview') &&
                        batchSpec.applyURL &&
                        batchSpec.state === BatchSpecState.FAILED && (
                            <Button
                                to={`${executionURL}/preview`}
                                data-tooltip="Execution didn't finish successfully in all workspaces. The batch spec might have less changeset specs than expected."
                                variant="warning"
                                outline={true}
                                as={Link}
                            >
                                <Icon className="mb-0 mr-2 text-warning" as={AlertCircleIcon} />
                                Preview
                            </Button>
                        )}
                </ButtonGroup>
                {/* TODO: Move me out to main page */}
                {cancelError && <ErrorAlert error={cancelError} />}
                {retryError && <ErrorAlert error={retryError} />}
            </span>
        </div>
    )
}

interface EditPageProps extends ThemeProps {
    name: string
    content: string
}

const EditPage: React.FunctionComponent<React.PropsWithChildren<EditPageProps>> = ({ name, content, isLightTheme }) => (
    <div className={classNames(styles.layoutContainer, 'h-100')}>
        <BatchSpec name={name} originalInput={content} isLightTheme={isLightTheme} className={styles.batchSpec} />
    </div>
)

const WORKSPACES_LIST_SIZE = 'batch-changes.ssbc-workspaces-list-size'

interface ExecutionPageProps extends ThemeProps, RouteComponentProps<{}> {
    batchSpec: BatchSpecExecutionFields
}

const ExecutionPage: React.FunctionComponent<React.PropsWithChildren<ExecutionPageProps>> = ({ match, ...props }) => (
    <Switch>
        <Route
            path={`${match.url}/workspaces/:workspaceID`}
            render={({
                match: {
                    params: { workspaceID },
                },
            }: RouteComponentProps<{ workspaceID: string }>) => (
                <ExecutionWorkspacesPage {...props} executionURL={match.url} selectedWorkspaceID={workspaceID} />
            )}
            exact={true}
        />
        <Route
            path={match.url}
            render={() => <ExecutionWorkspacesPage {...props} executionURL={match.url} />}
            exact={true}
        />
        <Route component={NotFoundPage} key="hardcoded-key" />
    </Switch>
)

interface ExecutionWorkspacesPageProps extends ThemeProps {
    executionURL: string
    batchSpec: BatchSpecExecutionFields
    selectedWorkspaceID?: string
}

const ExecutionWorkspacesPage: React.FunctionComponent<React.PropsWithChildren<ExecutionWorkspacesPageProps>> = ({
    batchSpec,
    selectedWorkspaceID,
    executionURL,
    isLightTheme,
}) => {
    const history = useHistory()
    const deselectWorkspace = useCallback(() => history.push(executionURL), [executionURL, history])

    return (
        <>
            {batchSpec.failureMessage && <ErrorAlert error={batchSpec.failureMessage} />}
            <div className={classNames(styles.layoutContainer, 'd-flex flex-1')}>
                <Panel defaultSize={500} minSize={405} maxSize={1400} position="left" storageKey={WORKSPACES_LIST_SIZE}>
                    <Workspaces
                        batchSpecID={batchSpec.id}
                        selectedNode={selectedWorkspaceID}
                        executionURL={executionURL}
                    />
                </Panel>
                <SelectedWorkspace
                    workspace={selectedWorkspaceID ?? null}
                    isLightTheme={isLightTheme}
                    deselectWorkspace={deselectWorkspace}
                />
            </div>
        </>
    )
}

interface PreviewPageProps extends TelemetryProps, ThemeProps {
    batchSpecID: Scalars['ID']
    batchSpec: BatchSpecExecutionFields
    authenticatedUser: AuthenticatedUser
}
const PreviewPage: React.FunctionComponent<React.PropsWithChildren<PreviewPageProps>> = ({
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

interface SelectedWorkspaceProps extends ThemeProps {
    deselectWorkspace: () => void
    workspace: Scalars['ID'] | null
}

const SelectedWorkspace: React.FunctionComponent<React.PropsWithChildren<SelectedWorkspaceProps>> = ({
    workspace,
    deselectWorkspace,
    isLightTheme,
}) => (
    <Card className="w-100 overflow-auto flex-grow-1">
        {/* This is necessary to prevent the margin collapse on `Card` */}
        <div className="w-100">
            <CardBody>
                {workspace ? (
                    <WorkspaceDetails
                        id={workspace}
                        isLightTheme={isLightTheme}
                        deselectWorkspace={deselectWorkspace}
                    />
                ) : (
                    <Typography.H3 className="text-center my-3">Select a workspace to view details.</Typography.H3>
                )}
            </CardBody>
        </div>
    </Card>
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)
