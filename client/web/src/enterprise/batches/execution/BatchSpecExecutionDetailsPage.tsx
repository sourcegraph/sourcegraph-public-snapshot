import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
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

            <h2>Timeline</h2>
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
