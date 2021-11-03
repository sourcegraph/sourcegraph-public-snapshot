import classNames from 'classnames'
import { parseISO } from 'date-fns/esm'
import { isArray, isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useEffect, useMemo, useReducer, useState } from 'react'
import { useHistory } from 'react-router'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecState, BatchSpecWorkspaceState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'
import { DiffStat } from '@sourcegraph/web/src/components/diff/DiffStat'
import { FileDiffConnection } from '@sourcegraph/web/src/components/diff/FileDiffConnection'
import { FileDiffNode } from '@sourcegraph/web/src/components/diff/FileDiffNode'
import { ExecutionLogEntry } from '@sourcegraph/web/src/components/ExecutionLogEntry'
import { FilteredConnectionQueryArguments } from '@sourcegraph/web/src/components/FilteredConnection'
import { LogOutput } from '@sourcegraph/web/src/components/LogOutput'
import { Timeline, TimelineStage } from '@sourcegraph/web/src/components/Timeline'
import { Container, PageHeader, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import {
    BatchSpecExecutionFields,
    BatchSpecWorkspaceFields,
    BatchSpecWorkspaceListFields,
    BatchSpecWorkspaceStepFields,
    Scalars,
} from '../../../graphql-operations'
import { BatchSpec } from '../BatchSpec'

import {
    cancelBatchSpecExecution,
    fetchBatchSpecExecution as _fetchBatchSpecExecution,
    fetchBatchSpecWorkspace,
    queryBatchSpecWorkspaceStepFileDiffs,
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
        <>
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
                                Batch Spec <BatchSpecStateBadge state={batchSpec.state} />
                            </>
                        ),
                    },
                ]}
                actions={
                    (batchSpec.state === BatchSpecState.QUEUED || batchSpec.state === BatchSpecState.PROCESSING) && (
                        <>
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
                        </>
                    )
                }
                className="mb-3"
            />

            {batchSpec.failureMessage && <ErrorAlert error={batchSpec.failureMessage} />}

            <h2>Input spec</h2>
            <Container className="mb-3">
                <BatchSpec originalInput={batchSpec.originalInput} className={styles.batchSpec} />
            </Container>
            <div className="d-flex justify-content-between mb-2">
                <h2 className="mb-0">Workspaces</h2>
                <div>
                    {batchSpec.startedAt && (
                        <>
                            Total run time:{' '}
                            <Duration start={batchSpec.startedAt} end={batchSpec.finishedAt ?? undefined} />
                        </>
                    )}
                </div>
            </div>
            <div className="row mb-3">
                <div className="col-4">
                    <WorkspacesList nodes={batchSpec.workspaceResolution!.workspaces.nodes} />
                </div>
                <div className="col-8">
                    <SelectedWorkspace workspace={selectedWorkspace} isLightTheme={isLightTheme} />
                </div>
            </div>

            {batchSpec.applyURL && (
                <>
                    <h2>Execution result</h2>
                    <div className="alert alert-info d-flex justify-space-between align-items-center">
                        <span className="flex-grow-1">Batch spec has been created.</span>
                        <Link to={batchSpec.applyURL} className="btn btn-primary">
                            Preview changes
                        </Link>
                    </div>
                </>
            )}
        </>
    )
}

const WorkspacesList: React.FunctionComponent<{ nodes: BatchSpecWorkspaceListFields[] }> = ({ nodes }) => (
    <div className="card">
        <ul className={classNames(styles.workspacesList, 'list-group list-group-flush')}>
            {nodes.map(workspaceNode => (
                <li className="list-group-item" key={workspaceNode.id}>
                    <div className={classNames(styles.workspaceRepo, 'd-flex justify-content-between mb-1')}>
                        <div>
                            <WorkspaceStateIcon state={workspaceNode.state} />{' '}
                            <Link to={`?workspace=${workspaceNode.id}`}>{workspaceNode.repository.name}</Link>
                        </div>
                        {workspaceNode.diffStat && <DiffStat {...workspaceNode.diffStat} expandedCounts={true} />}
                    </div>
                    <span className="badge badge-secondary">{workspaceNode.branch.name}</span>
                </li>
            ))}
        </ul>
    </div>
)

const SelectedWorkspace: React.FunctionComponent<{ workspace: Scalars['ID'] | null } & ThemeProps> = ({
    workspace,
    isLightTheme,
}) => {
    if (workspace === null) {
        return (
            <Container>
                <h3 className="text-center">Select workspace to get started</h3>
            </Container>
        )
    }
    return (
        <Container>
            <WorkspaceNode id={workspace} isLightTheme={isLightTheme} />
        </Container>
    )
}

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

const WorkspaceNode: React.FunctionComponent<
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
                <h4>
                    <WorkspaceStateIcon state={workspace.state} /> {workspace.repository.name}
                </h4>
                <div>
                    {workspace.startedAt && (
                        <Duration start={workspace.startedAt} end={workspace.finishedAt ?? undefined} />
                    )}
                    <button type="button" className="btn btn-outline-secondary btn-sm ml-2" onClick={onClose}>
                        <CloseIcon className="icon-inline" />
                    </button>
                </div>
            </div>
            {workspace.failureMessage && <ErrorAlert error={workspace.failureMessage} />}
            {typeof workspace.placeInQueue === 'number' && (
                <p>
                    <SyncIcon className="icon-inline text-muted" /> #{workspace.placeInQueue} in queue
                </p>
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
            <Collapsible
                title={<h4 className="mb-0">Timeline</h4>}
                titleClassName="flex-grow-1"
                defaultExpanded={false}
            >
                <ExecutionTimeline node={workspace} />
            </Collapsible>
        </>
    )
}

const WorkspaceStep: React.FunctionComponent<
    { step: BatchSpecWorkspaceStepFields; workspaceID: Scalars['ID']; stepIndex: number } & ThemeProps
> = ({ step, stepIndex, isLightTheme, workspaceID }) => (
    <Collapsible
        className="card mb-2"
        titleClassName="w-100"
        title={
            <div className="card-body">
                <div className="d-flex justify-content-between">
                    <div>
                        <StepStateIcon step={step} /> <strong>Step {stepIndex + 1}</strong>{' '}
                        <span className="text-monospace">{step.run.slice(0, 25)}...</span>
                        <StepTimer step={step} />
                    </div>
                    <div>{step.diffStat && <DiffStat {...step.diffStat} expandedCounts={true} />}</div>
                </div>
            </div>
        }
    >
        <div className="p-2">
            <Tabs size="medium">
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
                            {step.startedAt && step.outputLines && <LogOutput text={step.outputLines.join('\n')} />}
                            {!step.startedAt && <p className="text-muted">Step not started yet</p>}
                            {step.startedAt && !step.outputLines && <LogOutput text="_No logs.._" />}
                        </div>
                    </TabPanel>
                    <TabPanel key="output-variables">
                        <div className="p-2">
                            {!step.startedAt && <p className="text-muted">Step not started yet</p>}
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
                                <p className="text-muted">No environment variables specified</p>
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
                        <div className="p-2">
                            <p className="text-monospace">{step.run}</p>
                            <p className="text-monospace mb-0">{step.container}</p>
                        </div>
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </div>
    </Collapsible>
)

const WorkspaceStateIcon: React.FunctionComponent<{ state: BatchSpecWorkspaceState }> = ({ state }) => {
    switch (state) {
        case BatchSpecWorkspaceState.PENDING:
            return null
        case BatchSpecWorkspaceState.QUEUED:
            return <TimerSandIcon className="icon-inline text-muted" />
        case BatchSpecWorkspaceState.PROCESSING:
            return <LoadingSpinner className="icon-inline text-muted" />
        case BatchSpecWorkspaceState.SKIPPED:
            return <LinkVariantRemoveIcon className="icon-inline text-muted" />
        case BatchSpecWorkspaceState.CANCELED:
        case BatchSpecWorkspaceState.CANCELING:
        case BatchSpecWorkspaceState.FAILED:
            return <AlertCircleIcon className="icon-inline text-danger" />
        case BatchSpecWorkspaceState.COMPLETED:
            return <CheckCircleIcon className="icon-inline text-success" />
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
