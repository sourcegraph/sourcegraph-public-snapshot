import classNames from 'classnames'
import { parseISO } from 'date-fns/esm'
import { isArray, isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useEffect, useMemo, useReducer, useState } from 'react'
import { useHistory } from 'react-router'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecState, BatchSpecWorkspaceState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'
import { DiffStat } from '@sourcegraph/web/src/components/diff/DiffStat'
import { FileDiffConnection } from '@sourcegraph/web/src/components/diff/FileDiffConnection'
import { FileDiffNode } from '@sourcegraph/web/src/components/diff/FileDiffNode'
import { ExecutionLogEntry } from '@sourcegraph/web/src/components/ExecutionLogEntry'
import {
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { LogOutput } from '@sourcegraph/web/src/components/LogOutput'
import { Timeline, TimelineStage } from '@sourcegraph/web/src/components/Timeline'
import { Container, PageHeader, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import {
    BatchSpecExecutionFields,
    BatchSpecWorkspaceChangesetSpecFields,
    BatchSpecWorkspaceFields,
    BatchSpecWorkspaceListFields,
    BatchSpecWorkspaceStepFields,
    Scalars,
} from '../../../graphql-operations'
import { BatchSpec } from '../BatchSpec'
import { queryChangesetSpecFileDiffs } from '../preview/list/backend'
import { ChangesetSpecFileDiffConnection } from '../preview/list/ChangesetSpecFileDiffConnection'

import {
    cancelBatchSpecExecution,
    fetchBatchSpecExecution as _fetchBatchSpecExecution,
    fetchBatchSpecWorkspace,
    queryBatchSpecWorkspaces,
    queryBatchSpecWorkspaceStepFileDiffs,
    retryWorkspaceExecution,
} from './backend'
import styles from './BatchSpecExecutionDetailsPage.module.scss'

export interface BatchSpecExecutionDetailsPageProps extends ThemeProps {
    batchSpecID: Scalars['ID']

    /** For testing only. */
    fetchBatchSpecExecution?: typeof _fetchBatchSpecExecution
    /** For testing only. */
    expandStage?: string
}

export const BatchSpecExecutionDetailsPage: React.FunctionComponent<BatchSpecExecutionDetailsPageProps> = ({
    batchSpecID,
    isLightTheme,
    fetchBatchSpecExecution = _fetchBatchSpecExecution,
}) => {
    const [batchSpec, setBatchSpecExecution] = useState<BatchSpecExecutionFields | null | undefined>()

    useEffect(() => {
        const subscription = fetchBatchSpecExecution(batchSpecID)
            .pipe(
                repeatWhen(notifier => notifier.pipe(delay(2500))),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
            .subscribe(execution => {
                setBatchSpecExecution(execution)
            })

        return () => subscription.unsubscribe()
    }, [fetchBatchSpecExecution, batchSpecID])

    const history = useHistory()

    const selectedWorkspace = useMemo(() => {
        const query = new URLSearchParams(history.location.search)
        return query.get('workspace')
    }, [history.location.search])

    const [isCanceling, setIsCanceling] = useState<boolean | Error>(false)
    const cancelExecution = useCallback(async () => {
        try {
            const execution = await cancelBatchSpecExecution(batchSpecID)
            setBatchSpecExecution(execution)
        } catch (error) {
            setIsCanceling(asError(error))
        }
    }, [batchSpecID])

    // Is loading.
    if (batchSpec === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }

    // Is not found.
    if (batchSpec === null) {
        return <HeroPage icon={AlertCircleIcon} title="Execution not found" />
    }

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
                    {
                        text: (
                            <>
                                {batchSpec.description.name} <BatchSpecStateBadge state={batchSpec.state} />
                            </>
                        ),
                    },
                ]}
                actions={
                    <div className="d-flex">
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
                                <div className="mx-2 text-center text-muted">
                                    <h3 className="text-danger">
                                        {batchSpec.workspaceResolution.workspaces.stats.errored}
                                    </h3>
                                    Errors
                                </div>
                            </>
                        )}
                        {batchSpec.workspaceResolution?.workspaces.stats && (
                            <>
                                <div className="mx-2 text-center text-muted">
                                    <h3 className="text-success">
                                        {batchSpec.workspaceResolution.workspaces.stats.completed}
                                    </h3>
                                    Complete
                                </div>
                            </>
                        )}
                        {batchSpec.workspaceResolution?.workspaces.stats && (
                            <>
                                <div className="mx-2 text-center text-muted">
                                    <h3>{batchSpec.workspaceResolution.workspaces.stats.processing}</h3>
                                    Working
                                </div>
                            </>
                        )}
                        {batchSpec.workspaceResolution?.workspaces.stats && (
                            <>
                                <div className="mx-2 text-center text-muted">
                                    <h3>{batchSpec.workspaceResolution.workspaces.stats.queued}</h3>
                                    Queued
                                </div>
                            </>
                        )}
                        {batchSpec.workspaceResolution?.workspaces.stats && (
                            <>
                                <div className="mx-2 text-center text-muted">
                                    <h3>{batchSpec.workspaceResolution.workspaces.stats.ignored}</h3>
                                    Ignored
                                </div>
                            </>
                        )}
                        {(batchSpec.state === BatchSpecState.QUEUED ||
                            batchSpec.state === BatchSpecState.PROCESSING) && (
                            <span>
                                <button
                                    type="button"
                                    className="btn btn-outline-secondary"
                                    onClick={cancelExecution}
                                    disabled={isCanceling === true}
                                >
                                    {isCanceling !== true && <>Cancel</>}
                                    {isCanceling === true && (
                                        <>
                                            <LoadingSpinner className="icon-inline" /> Canceling
                                        </>
                                    )}
                                </button>
                                {isErrorLike(isCanceling) && <ErrorAlert error={isCanceling} />}
                            </span>
                        )}
                        {batchSpec.state === BatchSpecState.COMPLETED && batchSpec.applyURL && (
                            <span>
                                <Link to={batchSpec.applyURL} className="btn btn-primary">
                                    Preview
                                </Link>
                            </span>
                        )}
                    </div>
                }
                className="mb-3"
            />

            {batchSpec.failureMessage && <ErrorAlert error={batchSpec.failureMessage} />}

            <Collapsible title="Input spec" wholeTitleClickable={false} titleClassName="flex-grow-1" className="mb-3">
                <BatchSpec originalInput={batchSpec.originalInput} className={styles.batchSpec} />
            </Collapsible>
            <div className="d-flex mb-3">
                <div className={styles.workspacesListContainer}>
                    <h3 className="mb-2">Workspaces</h3>
                    <WorkspacesList batchSpecID={batchSpec.id} selectedNode={selectedWorkspace ?? undefined} />
                </div>
                <div className="flex-grow-1">
                    <SelectedWorkspace workspace={selectedWorkspace} isLightTheme={isLightTheme} />
                </div>
            </div>
        </div>
    )
}

const WorkspacesList: React.FunctionComponent<{
    batchSpecID: Scalars['ID']
    selectedNode?: Scalars['ID']
}> = ({ batchSpecID, selectedNode }) => {
    const query = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryBatchSpecWorkspaces({
                node: batchSpecID,
                first: args.first ?? null,
                after: args.after ?? null,
            }).pipe(
                repeatWhen(notifier => notifier.pipe(delay(1000))),
                distinctUntilChanged((a, b) => isEqual(a, b))
            ),
        [batchSpecID]
    )
    const history = useHistory()
    return (
        <FilteredConnection<BatchSpecWorkspaceListFields, Omit<WorkspaceNodeProps, 'node'>>
            history={history}
            location={history.location}
            nodeComponent={WorkspaceNode}
            nodeComponentProps={{ selectedNode }}
            queryConnection={query}
            hideSearch={true}
            defaultFirst={20}
            noun="workspace"
            pluralNoun="workspaces"
            listClassName={classNames(styles.workspacesList, 'list-group list-group-flush')}
            listComponent="ul"
            withCenteredSummary={true}
            cursorPaging={true}
            noSummaryIfAllNodesVisible={true}
        />
    )
}

interface WorkspaceNodeProps {
    node: BatchSpecWorkspaceListFields
    selectedNode?: Scalars['ID']
}

const WorkspaceNode: React.FunctionComponent<WorkspaceNodeProps> = ({ node, selectedNode }) => (
    <li className={classNames('list-group-item', node.id === selectedNode && styles.workspaceSelected)}>
        <Link to={`?workspace=${node.id}`}>
            <div className={classNames(styles.workspaceRepo, 'd-flex justify-content-between mb-1')}>
                <WorkspaceStateIcon
                    cachedResultFound={node.cachedResultFound}
                    state={node.state}
                    className={classNames(styles.workspaceListIcon, 'mr-2')}
                />
                <strong className={classNames(styles.workspaceName, 'flex-grow-1')}>{node.repository.name}</strong>
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} />}
            </div>
            {/* Only display the branch if it's not the default branch. */}
            {node.repository.defaultBranch?.abbrevName !== node.branch.abbrevName && (
                <span className="badge badge-secondary">{node.branch.abbrevName}</span>
            )}
        </Link>
    </li>
)

const SelectedWorkspace: React.FunctionComponent<{ workspace: Scalars['ID'] | null } & ThemeProps> = ({
    workspace,
    isLightTheme,
}) => {
    if (workspace === null) {
        return (
            <Container>
                <h3 className="text-center mb-0">Select a workspace to view details.</h3>
            </Container>
        )
    }
    return (
        <Container>
            <WorkspaceDetails id={workspace} isLightTheme={isLightTheme} />
        </Container>
    )
}

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

const WorkspaceDetails: React.FunctionComponent<
    {
        id: Scalars['ID']
    } & ThemeProps
> = ({ id, isLightTheme }) => {
    const history = useHistory()
    const onClose = useCallback(() => {
        history.push(history.location.pathname)
    }, [history])

    // Fetch and poll latest workspace information.
    const workspace = useObservable(
        useMemo(() => fetchBatchSpecWorkspace(id).pipe(repeatWhen(notifier => notifier.pipe(delay(2500)))), [id])
    )

    const [retrying, setRetrying] = useState<boolean | ErrorLike>(false)
    const onRetry = useCallback(async () => {
        setRetrying(true)
        try {
            await retryWorkspaceExecution(id)
            setRetrying(false)
        } catch (error) {
            setRetrying(asError(error))
        }
    }, [id])

    if (workspace === undefined) {
        return <LoadingSpinner />
    }

    if (workspace === null) {
        return <NotFoundPage />
    }

    if (isErrorLike(workspace)) {
        return <ErrorAlert error={workspace} />
    }
    return (
        <>
            <div className="d-flex justify-content-between">
                <h3>
                    <WorkspaceStateIcon cachedResultFound={workspace.cachedResultFound} state={workspace.state} />{' '}
                    {workspace.repository.name}
                </h3>
                <button type="button" className="btn btn-link btn-sm p-0 ml-2" onClick={onClose}>
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
            <div className="text-muted mb-3">
                {workspace.path && <>{workspace.path} | </>}
                <SourceBranchIcon className="icon-inline" /> {workspace.branch.abbrevName}
                {workspace.startedAt && (
                    <>
                        {' '}
                        | Total time:{' '}
                        <strong>
                            <Duration start={workspace.startedAt} end={workspace.finishedAt ?? undefined} />
                        </strong>
                    </>
                )}
            </div>
            {workspace.failureMessage && (
                <>
                    <ErrorAlert error={workspace.failureMessage} className="mb-2" />
                    <button
                        type="button"
                        className="btn btn-outline-danger mb-3"
                        onClick={onRetry}
                        disabled={retrying === true}
                    >
                        <SyncIcon className="icon-inline" /> Retry
                    </button>
                    {isErrorLike(retrying) && <ErrorAlert error={retrying} />}
                </>
            )}
            {typeof workspace.placeInQueue === 'number' && (
                <p>
                    <SyncIcon className="icon-inline text-muted" /> #{workspace.placeInQueue} in queue
                </p>
            )}
            {workspace.state === BatchSpecWorkspaceState.COMPLETED && (
                <>
                    <h4>Changeset specs</h4>
                    {workspace.changesetSpecs?.length === 0 && (
                        <p className="mb-2 text-muted">This workspace generated no changeset specs.</p>
                    )}
                    {workspace.changesetSpecs?.map(changesetSpec => (
                        <ChangesetSpecNode key={changesetSpec.id} node={changesetSpec} isLightTheme={isLightTheme} />
                    ))}
                </>
            )}
            <h4>Steps</h4>
            {workspace.steps.map((step, index) => (
                <WorkspaceStep
                    step={step}
                    stepIndex={index}
                    workspaceID={workspace.id}
                    key={index}
                    isLightTheme={isLightTheme}
                />
            ))}
            {!workspace.cachedResultFound && workspace.state !== BatchSpecWorkspaceState.SKIPPED && (
                <Collapsible
                    title={<h4 className="mb-0">Timeline</h4>}
                    titleClassName="flex-grow-1"
                    defaultExpanded={false}
                    className="mt-2"
                >
                    <ExecutionTimeline node={workspace} />
                </Collapsible>
            )}
        </>
    )
}

const ChangesetSpecNode: React.FunctionComponent<{ node: BatchSpecWorkspaceChangesetSpecFields } & ThemeProps> = ({
    node,
    isLightTheme,
}) => {
    const history = useHistory()
    // TODO: This should not happen. When the workspace is visibile, the changeset spec should be visible as well.
    if (node.__typename === 'HiddenChangesetSpec') {
        return (
            <div className="card">
                <div className="card-body">
                    <h4>Changeset in a hidden repo</h4>
                </div>
            </div>
        )
    }
    // This should not happen.
    if (node.description.__typename === 'ExistingChangesetReference') {
        return null
    }
    return (
        <div className="card mb-2">
            <div className="card-body">
                <div className="d-flex justify-content-between">
                    <h4>{node.description.title}</h4>
                    <DiffStat {...node.description.diffStat} expandedCounts={true} />
                </div>
                <p>
                    <Link to={node.description.baseRepository.url}>{node.description.baseRepository.name}</Link>
                </p>
                <p>
                    <span className="badge badge-secondary">{node.description.baseRef}</span> &larr;
                    <span className="badge badge-secondary">{node.description.headRef}</span>
                </p>
                <p>
                    <strong>Published:</strong> <PublishedValue published={node.description.published} />
                </p>
                <Collapsible
                    title={<h4 className="mb-0">Changed files</h4>}
                    titleClassName="flex-grow-1"
                    defaultExpanded={false}
                >
                    <ChangesetSpecFileDiffConnection
                        history={history}
                        isLightTheme={isLightTheme}
                        location={history.location}
                        spec={node.id}
                        queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                    />
                </Collapsible>
            </div>
        </div>
    )
}

const PublishedValue: React.FunctionComponent<{ published: Scalars['PublishedValue'] | null }> = ({ published }) => {
    if (published === null) {
        return <i>select from UI when applying</i>
    }
    if (published === 'draft') {
        return <>draft</>
    }
    return <>{String(published)}</>
}

const WorkspaceStep: React.FunctionComponent<
    { step: BatchSpecWorkspaceStepFields; workspaceID: Scalars['ID']; stepIndex: number } & ThemeProps
> = ({ step, stepIndex, isLightTheme, workspaceID }) => {
    const outputLines = useMemo(() => {
        const outputLines = step.outputLines
        if (outputLines !== null) {
            if (
                outputLines.every(
                    line =>
                        line
                            .replaceAll(/'^std(out|err):'/g, '')
                            .replaceAll('\n', '')
                            .trim() === ''
                )
            ) {
                outputLines.push('stderr: This command did not produce any logs')
            }
            if (step.exitCode !== null) {
                outputLines.push(`\nstdout: \nstdout: Command exited with status ${step.exitCode}`)
            }
        }
        return outputLines
    }, [step.exitCode, step.outputLines])

    return (
        <Collapsible
            className="card mb-2"
            titleClassName="w-100"
            title={
                <div className="card-body">
                    <div className="d-flex justify-content-between">
                        <div className="flex-grow-1">
                            <StepStateIcon step={step} /> <strong>Step {stepIndex + 1}</strong>{' '}
                            <span className="text-monospace">{step.run.slice(0, 25)}...</span>
                        </div>
                        <div>{step.diffStat && <DiffStat {...step.diffStat} expandedCounts={true} />}</div>
                        <span className="text-monospace ml-2">
                            <StepTimer step={step} />
                        </span>
                    </div>
                </div>
            }
        >
            <div className="p-2">
                {!step.skipped && (
                    <Tabs size="small">
                        <TabList>
                            <Tab key="logs">Logs</Tab>
                            <Tab key="output-variables">Output variables</Tab>
                            <Tab key="diff">Diff</Tab>
                            <Tab key="files-env">Files / Env</Tab>
                            <Tab key="command-container">Commands / container</Tab>
                        </TabList>
                        <TabPanels>
                            <TabPanel key="logs">
                                <div className="p-2">
                                    {!step.startedAt && <p className="text-muted">Step not started yet</p>}
                                    {step.startedAt && outputLines && <LogOutput text={outputLines.join('\n')} />}
                                </div>
                            </TabPanel>
                            <TabPanel key="output-variables">
                                <div className="p-2">
                                    {!step.startedAt && <p className="text-muted">Step not started yet</p>}
                                    {step.environment.length === 0 && (
                                        <p className="text-muted mb-0">No output variables specified</p>
                                    )}
                                    <ul>
                                        {step.outputVariables?.map(variable => (
                                            <li key={variable.name}>
                                                {variable.name}: {variable.value}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            </TabPanel>
                            <TabPanel key="diff">
                                <div className="p-2">
                                    {!step.startedAt && <p className="text-muted">Step not started yet</p>}
                                    {step.startedAt && (
                                        <WorkspaceStepFileDiffConnection
                                            isLightTheme={isLightTheme}
                                            step={stepIndex + 1}
                                            workspaceID={workspaceID}
                                        />
                                    )}
                                </div>
                            </TabPanel>
                            <TabPanel key="files-env">
                                <div className="p-2">
                                    {step.environment.length === 0 && (
                                        <p className="text-muted mb-0">No environment variables specified</p>
                                    )}
                                    <ul>
                                        {step.environment.map(variable => (
                                            <li key={variable.name}>
                                                {variable.name}: {variable.value}
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            </TabPanel>
                            <TabPanel key="command-container">
                                <div className="p-2 pb-0">
                                    <LogOutput text={step.run} className="mb-2" />
                                    <p className="mb-0">
                                        Using container <span className="text-monospace">{step.container}</span>
                                    </p>
                                </div>
                            </TabPanel>
                        </TabPanels>
                    </Tabs>
                )}
                {step.skipped && (
                    <p className="mb-0">
                        <strong>Step has been skipped.</strong>
                    </p>
                )}
            </div>
        </Collapsible>
    )
}

const WorkspaceStateIcon: React.FunctionComponent<{
    cachedResultFound: boolean
    state: BatchSpecWorkspaceState
    className?: string
}> = ({ cachedResultFound, state, className }) => {
    switch (state) {
        case BatchSpecWorkspaceState.PENDING:
            return null
        case BatchSpecWorkspaceState.QUEUED:
            return <TimerSandIcon className={classNames('icon-inline text-muted', className)} />
        case BatchSpecWorkspaceState.PROCESSING:
            return <LoadingSpinner className={classNames('icon-inline text-muted', className)} />
        case BatchSpecWorkspaceState.SKIPPED:
            return <LinkVariantRemoveIcon className={classNames('icon-inline text-muted', className)} />
        case BatchSpecWorkspaceState.CANCELED:
        case BatchSpecWorkspaceState.CANCELING:
        case BatchSpecWorkspaceState.FAILED:
            return <AlertCircleIcon className={classNames('icon-inline text-danger', className)} />
        case BatchSpecWorkspaceState.COMPLETED:
            if (cachedResultFound) {
                return <ContentSaveIcon className={classNames('icon-inline text-success', className)} />
            }
            return <CheckCircleIcon className={classNames('icon-inline text-success', className)} />
    }
}

const StepStateIcon: React.FunctionComponent<{ step: BatchSpecWorkspaceStepFields }> = ({ step }) => {
    if (step.skipped) {
        return <LinkVariantRemoveIcon className="icon-inline text-muted" />
    }
    if (!step.startedAt) {
        return <TimerSandIcon className="icon-inline text-muted" />
    }
    if (!step.finishedAt) {
        return <LoadingSpinner className="icon-inline text-muted" />
    }
    if (step.exitCode === 0) {
        return <CheckCircleIcon className="icon-inline text-success" />
    }
    return <AlertCircleIcon className="icon-inline text-danger" />
}

const StepTimer: React.FunctionComponent<{ step: BatchSpecWorkspaceStepFields }> = ({ step }) => {
    if (!step.startedAt) {
        return null
    }
    return <Duration start={step.startedAt} end={step.finishedAt ?? undefined} />
}

interface ExecutionTimelineProps {
    node: BatchSpecWorkspaceFields
    className?: string

    /** For testing only. */
    now?: () => Date
    expandStage?: string
}

const ExecutionTimeline: React.FunctionComponent<ExecutionTimelineProps> = ({ node, className, now, expandStage }) => {
    const stages = useMemo(
        () => [
            { icon: <TimerSandIcon />, text: 'Queued', date: node.queuedAt, className: 'bg-success' },
            {
                icon: <CheckIcon />,
                text: 'Began processing',
                date: node.startedAt,
                className: 'bg-success',
            },

            setupStage(node, expandStage === 'setup', now),
            batchPreviewStage(node, expandStage === 'srcPreview', now),
            teardownStage(node, expandStage === 'teardown', now),

            node.state === BatchSpecWorkspaceState.COMPLETED
                ? { icon: <CheckIcon />, text: 'Finished', date: node.finishedAt, className: 'bg-success' }
                : node.state === BatchSpecWorkspaceState.CANCELED
                ? { icon: <AlertCircleIcon />, text: 'Canceled', date: node.finishedAt, className: 'bg-secondary' }
                : { icon: <AlertCircleIcon />, text: 'Failed', date: node.finishedAt, className: 'bg-danger' },
        ],
        [expandStage, node, now]
    )
    return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
}

const setupStage = (
    execution: BatchSpecWorkspaceFields,
    expand: boolean,
    now?: () => Date
): TimelineStage | undefined => {
    if (execution.stages === null) {
        return undefined
    }
    return execution.stages.setup.length === 0
        ? undefined
        : {
              text: 'Setup',
              details: execution.stages.setup.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(execution.stages.setup, expand),
          }
}

const batchPreviewStage = (
    execution: BatchSpecWorkspaceFields,
    expand: boolean,
    now?: () => Date
): TimelineStage | undefined => {
    if (execution.stages === null) {
        return undefined
    }
    return !execution.stages.srcExec
        ? undefined
        : {
              text: 'Create batch spec preview',
              details: (
                  <ExecutionLogEntry key={execution.stages.srcExec.key} logEntry={execution.stages.srcExec} now={now} />
              ),
              ...genericStage(execution.stages.srcExec, expand),
          }
}

const teardownStage = (
    execution: BatchSpecWorkspaceFields,
    expand: boolean,
    now?: () => Date
): TimelineStage | undefined => {
    if (execution.stages === null) {
        return undefined
    }
    return execution.stages.teardown.length === 0
        ? undefined
        : {
              text: 'Teardown',
              details: execution.stages.teardown.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(execution.stages.teardown, expand),
          }
}

const genericStage = <E extends { startTime: string; exitCode: number | null }>(
    value: E | E[],
    expand: boolean
): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expanded'> => {
    const finished = isArray(value) ? value.every(logEntry => logEntry.exitCode !== null) : value.exitCode !== null
    const success = isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: !finished ? <ProgressClockIcon /> : success ? <CheckIcon /> : <AlertCircleIcon />,
        date: isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expanded: expand || !(success || !finished),
    }
}

const Duration: React.FunctionComponent<{ start: Date | string; end?: Date | string }> = ({ start, end }) => {
    const startDate = typeof start === 'string' ? parseISO(start) : start
    const endDate = typeof end === 'string' ? parseISO(end) : end || new Date()
    let duration = endDate.getTime() / 1000 - startDate.getTime() / 1000
    const hours = Math.floor(duration / (60 * 60))
    duration -= hours * 60 * 60
    const minutes = Math.floor(duration / 60)
    duration -= minutes * 60
    const seconds = Math.floor(duration)

    const [, forceUpdate] = useReducer((any: number) => any + 1, 0)

    useEffect(() => {
        if (end === undefined) {
            const timer = setInterval(() => {
                forceUpdate()
            }, 1000)
            return () => {
                clearInterval(timer)
            }
        }
        return undefined
    }, [end])

    return (
        <>
            {leading0(hours)}:{leading0(minutes)}:{leading0(seconds)}
        </>
    )
}

function leading0(index: number): string {
    if (index < 10) {
        return '0' + String(index)
    }
    return String(index)
}

const WorkspaceStepFileDiffConnection: React.FunctionComponent<
    {
        workspaceID: Scalars['ID']
        step: number
    } & ThemeProps
> = ({ workspaceID, step, isLightTheme }) => {
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryBatchSpecWorkspaceStepFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                node: workspaceID,
                step,
            }),
        [workspaceID, step]
    )
    const history = useHistory()
    return (
        <FileDiffConnection
            listClassName="list-group list-group-flush"
            noun="changed file"
            pluralNoun="changed files"
            queryConnection={queryFileDiffs}
            nodeComponent={FileDiffNode}
            nodeComponentProps={{
                history,
                location: history.location,
                isLightTheme,
                persistLines: true,
                lineNumbers: true,
            }}
            defaultFirst={15}
            hideSearch={true}
            noSummaryIfAllNodesVisible={true}
            history={history}
            location={history.location}
            useURLQuery={false}
            cursorPaging={true}
        />
    )
}

const BatchSpecStateBadge: React.FunctionComponent<{ state: BatchSpecState }> = ({ state }) => {
    switch (state) {
        case BatchSpecState.PENDING:
        case BatchSpecState.QUEUED:
        case BatchSpecState.PROCESSING:
        case BatchSpecState.CANCELED:
        case BatchSpecState.CANCELING:
            return <span className="badge badge-secondary">{state}</span>
        case BatchSpecState.FAILED:
            return <span className="badge badge-danger">{state}</span>
        case BatchSpecState.COMPLETED:
            return <span className="badge badge-success">{state}</span>
    }
}
