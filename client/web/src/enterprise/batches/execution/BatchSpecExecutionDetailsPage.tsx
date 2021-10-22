import { parseISO } from 'date-fns/esm'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useEffect, useMemo, useReducer, useState } from 'react'
import { useHistory } from 'react-router'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecState, BatchSpecWorkspaceState } from '@sourcegraph/shared/src/graphql-operations'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'
import { DiffStat } from '@sourcegraph/web/src/components/diff/DiffStat'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { BatchSpecExecutionFields, Scalars } from '../../../graphql-operations'
import { BatchSpec } from '../BatchSpec'

import { cancelBatchSpecExecution, fetchBatchSpecExecution as _fetchBatchSpecExecution } from './backend'

export interface BatchSpecExecutionDetailsPageProps {
    executionID: Scalars['ID']

    /** For testing only. */
    fetchBatchSpecExecution?: typeof _fetchBatchSpecExecution
    /** For testing only. */
    now?: () => Date
    /** For testing only. */
    expandStage?: string
}

export const BatchSpecExecutionDetailsPage: React.FunctionComponent<BatchSpecExecutionDetailsPageProps> = ({
    executionID,
    // now = () => new Date(),
    fetchBatchSpecExecution = _fetchBatchSpecExecution,
    // expandStage,
}) => {
    const [batchSpecExecution, setBatchSpecExecution] = useState<BatchSpecExecutionFields | null | undefined>()

    useEffect(() => {
        const subscription = fetchBatchSpecExecution(executionID)
            .pipe(
                repeatWhen(notifier => notifier.pipe(delay(2500))),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
            .subscribe(execution => {
                setBatchSpecExecution(execution)
            })

        return () => subscription.unsubscribe()
    }, [fetchBatchSpecExecution, executionID])

    const history = useHistory()

    const selectedNamespace = useMemo(() => {
        const query = new URLSearchParams(history.location.search)
        return query.get('workspace')
    }, [history.location.search])

    const [isCanceling, setIsCanceling] = useState<boolean | Error>(false)
    const cancelExecution = useCallback(async () => {
        try {
            const execution = await cancelBatchSpecExecution(executionID)
            setBatchSpecExecution(execution)
        } catch (error) {
            setIsCanceling(asError(error))
        }
    }, [executionID])

    // Is loading.
    if (batchSpecExecution === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }

    // Is not found.
    if (batchSpecExecution === null) {
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
                        to: `${batchSpecExecution.namespace.url}/batch-changes`,
                        text: batchSpecExecution.namespace.namespaceName,
                    },
                    {
                        text: (
                            <>
                                Execution <span className="badge badge-secondary">{batchSpecExecution.state}</span>
                            </>
                        ),
                    },
                ]}
                actions={
                    (batchSpecExecution.state === BatchSpecState.QUEUED ||
                        batchSpecExecution.state === BatchSpecState.PROCESSING) && (
                        <>
                            <button
                                type="button"
                                className="btn btn-outline-secondary"
                                onClick={cancelExecution}
                                disabled={isCanceling === true}
                            >
                                Cancel
                            </button>
                            {isErrorLike(isCanceling) && <ErrorAlert error={isCanceling} />}
                        </>
                    )
                }
                className="mb-3"
            />

            {batchSpecExecution.failureMessage && <ErrorAlert error={batchSpecExecution.failureMessage} />}

            <h2>Input spec</h2>
            <Container className="mb-3">
                <BatchSpec originalInput={batchSpecExecution.originalInput} />
            </Container>
            <div>
                {batchSpecExecution.startedAt && (
                    <Duration start={batchSpecExecution.startedAt} end={batchSpecExecution.finishedAt ?? undefined} />
                )}
            </div>
            <div className="row mb-3">
                <div className="col-4">
                    <h2>Workspaces</h2>
                    <Container>
                        <ul className="list-group">
                            {batchSpecExecution.workspaceResolution!.workspaces.nodes.map(workspaceNode => (
                                <li className="list-group-item" key={workspaceNode.id}>
                                    <WorkspaceStateIcon node={workspaceNode} />{' '}
                                    <Link to={`?workspace=${workspaceNode.id}`}>{workspaceNode.repository.name}</Link>
                                </li>
                            ))}
                        </ul>
                    </Container>
                </div>
                <div className="col-8">
                    <Container>
                        {selectedNamespace === null && <h3 className="text-center">Select workspace to get started</h3>}
                        {selectedNamespace !== null && (
                            <WorkspaceNode
                                node={
                                    batchSpecExecution.workspaceResolution!.workspaces.nodes.find(
                                        node => node.id === selectedNamespace
                                    )!
                                }
                            />
                        )}
                    </Container>
                </div>
            </div>
            {/* <ExecutionTimeline execution={batchSpecExecution} now={now} expandStage={expandStage} className="mb-3" /> */}

            {batchSpecExecution.applyURL && (
                <>
                    <h2>Execution result</h2>
                    <div className="alert alert-info d-flex justify-space-between align-items-center">
                        <span className="flex-grow-1">Batch spec has been created.</span>
                        <Link to={batchSpecExecution.applyURL} className="btn btn-primary">
                            Preview changes
                        </Link>
                    </div>
                </>
            )}
        </>
    )
}

type Workspace = NonNullable<BatchSpecExecutionFields['workspaceResolution']>['workspaces']['nodes'][0]
type Step = Workspace['steps'][0]

const WorkspaceNode: React.FunctionComponent<{
    node: Workspace
}> = ({ node }) => {
    const a = ''
    return (
        <>
            <h4>{node.repository.name}</h4>
            <p>Started at: {node.startedAt ?? 'not yet'}</p>
            <p>Finished at: {node.finishedAt ?? 'not yet'}</p>
            {node.failureMessage && <ErrorAlert error={node.failureMessage} />}
            <p>
                <b>Steps</b>
            </p>
            {node.steps.map((step, index) => (
                <Collapsible
                    key={index}
                    className="card"
                    titleClassName="w-100"
                    title={
                        <div className="card-body">
                            <div className="d-flex justify-content-between">
                                <div>
                                    <StepStateIcon step={step} />
                                    <strong>Step {index + 1}</strong>{' '}
                                    <span className="text-monospace">{step.run.slice(0, 25)}...</span>
                                    <StepTimer step={step} />
                                </div>
                                <div>{step.diffStat && <DiffStat {...step.diffStat} expandedCounts={true} />}</div>
                            </div>
                        </div>
                    }
                >
                    {step.container}
                    {step.outputLines && <div className="card">{step.outputLines.join('\n')}</div>}
                    <p>
                        <strong>Output variables:</strong>
                        <br />
                        {JSON.stringify(step.outputVariables)}
                    </p>
                </Collapsible>
            ))}
        </>
    )
}

const WorkspaceStateIcon: React.FunctionComponent<{ node: Workspace }> = ({ node }) => {
    switch (node.state) {
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
            return <ErrorIcon className="icon-inline text-danger" />
        case BatchSpecWorkspaceState.COMPLETED:
            return <CheckCircleIcon className="icon-inline text-success" />
    }
}

const StepStateIcon: React.FunctionComponent<{ step: Step }> = ({ step }) => {
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
    return <ErrorIcon className="icon-inline text-danger" />
}

const StepTimer: React.FunctionComponent<{ step: Step }> = ({ step }) => {
    if (!step.startedAt) {
        return null
    }
    return <Duration start={step.startedAt} end={step.finishedAt ?? undefined} />
}

// interface ExecutionTimelineProps {
//     execution: BatchSpecExecutionFields
//     className?: string

//     /** For testing only. */
//     now?: () => Date
//     expandStage?: string
// }

// const ExecutionTimeline: React.FunctionComponent<ExecutionTimelineProps> = ({
//     execution,
//     className,
//     now,
//     expandStage,
// }) => {
//     const stages = useMemo(
//         () => [
//             { icon: <TimerSandIcon />, text: 'Queued', date: execution.createdAt, className: 'bg-success' },
//             {
//                 icon: <CheckIcon />,
//                 text: 'Began processing',
//                 date: execution.startedAt,
//                 className: 'bg-success',
//             },

//             setupStage(execution, expandStage === 'setup', now),
//             batchPreviewStage(execution, expandStage === 'srcPreview', now),
//             teardownStage(execution, expandStage === 'teardown', now),

//             execution.state === BatchSpecState.COMPLETED
//                 ? { icon: <CheckIcon />, text: 'Finished', date: execution.finishedAt, className: 'bg-success' }
//                 : execution.state === BatchSpecState.CANCELED
//                 ? { icon: <ErrorIcon />, text: 'Canceled', date: execution.finishedAt, className: 'bg-secondary' }
//                 : { icon: <ErrorIcon />, text: 'Failed', date: execution.finishedAt, className: 'bg-danger' },
//         ],
//         [execution, now, expandStage]
//     )
//     return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
// }

// const setupStage = (
//     execution: BatchSpecExecutionFields,
//     expand: boolean,
//     now?: () => Date
// ): TimelineStage | undefined =>
//     execution.steps.setup.length === 0
//         ? undefined
//         : {
//               text: 'Setup',
//               details: execution.steps.setup.map(logEntry => (
//                   <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
//               )),
//               ...genericStage(execution.steps.setup, expand),
//           }

// const batchPreviewStage = (
//     execution: BatchSpecExecutionFields,
//     expand: boolean,
//     now?: () => Date
// ): TimelineStage | undefined =>
//     !execution.steps.srcPreview
//         ? undefined
//         : {
//               text: 'Create batch spec preview',
//               details: (
//                   <ExecutionLogEntry logEntry={execution.steps.srcPreview} now={now}>
//                       {execution.steps.srcPreview.out && <ParsedJsonOutput out={execution.steps.srcPreview.out} />}
//                   </ExecutionLogEntry>
//               ),
//               ...genericStage(execution.steps.srcPreview, expand),
//           }

// const teardownStage = (
//     execution: BatchSpecExecutionFields,
//     expand: boolean,
//     now?: () => Date
// ): TimelineStage | undefined =>
//     execution.steps.teardown.length === 0
//         ? undefined
//         : {
//               text: 'Teardown',
//               details: execution.steps.teardown.map(logEntry => (
//                   <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
//               )),
//               ...genericStage(execution.steps.teardown, expand),
//           }

// const genericStage = <E extends { startTime: string; exitCode: number | null }>(
//     value: E | E[],
//     expand: boolean
// ): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expanded'> => {
//     const finished = isArray(value) ? value.every(logEntry => logEntry.exitCode !== null) : value.exitCode !== null
//     const success = isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

//     return {
//         icon: !finished ? <ProgressClockIcon /> : success ? <CheckIcon /> : <ErrorIcon />,
//         date: isArray(value) ? value[0].startTime : value.startTime,
//         className: success || !finished ? 'bg-success' : 'bg-danger',
//         expanded: expand || !(success || !finished),
//     }
// }

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
