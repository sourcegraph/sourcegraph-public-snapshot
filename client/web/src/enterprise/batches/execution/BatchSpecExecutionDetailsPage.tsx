import { parseISO } from 'date-fns'
import { isArray, isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecExecutionState } from '@sourcegraph/shared/src/graphql-operations'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { ExecutionLogEntry, LogOutput } from '../../../components/ExecutionLogEntry'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { Timeline, TimelineStage } from '../../../components/Timeline'
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
    now = () => new Date(),
    fetchBatchSpecExecution = _fetchBatchSpecExecution,
    expandStage,
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
                    (batchSpecExecution.state === BatchSpecExecutionState.QUEUED ||
                        batchSpecExecution.state === BatchSpecExecutionState.PROCESSING) && (
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

            {batchSpecExecution.failure && <ErrorAlert error={batchSpecExecution.failure} />}

            <h2>Input spec</h2>
            <Container className="mb-3">
                <BatchSpec originalInput={batchSpecExecution.inputSpec} />
            </Container>

            <h2>Timeline</h2>
            <ExecutionTimeline execution={batchSpecExecution} now={now} expandStage={expandStage} className="mb-3" />

            {batchSpecExecution.batchSpec && (
                <>
                    <h2>Execution result</h2>
                    <div className="alert alert-info d-flex justify-space-between align-items-center">
                        <span className="flex-grow-1">Batch spec has been created.</span>
                        <Link to={batchSpecExecution.batchSpec.applyURL} className="btn btn-primary">
                            Preview changes
                        </Link>
                    </div>
                </>
            )}
        </>
    )
}

interface ExecutionTimelineProps {
    execution: BatchSpecExecutionFields
    className?: string

    /** For testing only. */
    now?: () => Date
    expandStage?: string
}

const ExecutionTimeline: React.FunctionComponent<ExecutionTimelineProps> = ({
    execution,
    className,
    now,
    expandStage,
}) => {
    const stages = useMemo(
        () => [
            { icon: <TimerSandIcon />, text: 'Queued', date: execution.createdAt, className: 'bg-success' },
            {
                icon: <CheckIcon />,
                text: 'Began processing',
                date: execution.startedAt,
                className: 'bg-success',
            },

            setupStage(execution, expandStage === 'setup', now),
            ...batchPreviewStages(execution, expandStage === 'srcPreview', now),
            teardownStage(execution, expandStage === 'teardown', now),

            execution.state === BatchSpecExecutionState.COMPLETED
                ? { icon: <CheckIcon />, text: 'Finished', date: execution.finishedAt, className: 'bg-success' }
                : execution.state === BatchSpecExecutionState.CANCELED
                ? { icon: <ErrorIcon />, text: 'Canceled', date: execution.finishedAt, className: 'bg-secondary' }
                : { icon: <ErrorIcon />, text: 'Failed', date: execution.finishedAt, className: 'bg-danger' },
        ],
        [execution, now, expandStage]
    )
    return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
}

const setupStage = (
    execution: BatchSpecExecutionFields,
    expand: boolean,
    now?: () => Date
): TimelineStage | undefined =>
    execution.steps.setup.length === 0
        ? undefined
        : {
              text: 'Setup',
              details: execution.steps.setup.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(execution.steps.setup, expand),
          }

const batchPreviewStages = (
    execution: BatchSpecExecutionFields,
    expand: boolean,
    now?: () => Date
): TimelineStage[] => {
    if (!execution.steps.srcPreview) {
        return []
    }

    const result: TimelineStage[] = []

    const parsed = parseLogLines(execution.steps.srcPreview.out)
    for (const operation of Object.keys(JSONLogLineOperation)) {
        if (operation === JSONLogLineOperation.EXECUTING_TASKS) {
            const possibleTuple = findLogLineTuple(parsed, operation)
            if (!possibleTuple) {
                continue
            }
            const [start] = possibleTuple as [ExecutingTasksJSONLogLine, ExecutingTasksJSONLogLine | undefined]

            const tasks = start.metadata.tasks

            const step = transformLogsForOperationToStage(
                parsed,
                operation as JSONLogLineOperation,
                expand,
                now,
                <ul className="list-group mt-3">
                    {tasks === null && <li className="list-group-item text-muted">No tasks</li>}
                    {tasks?.map(task => {
                        const tuple = findLogLineTuple(
                            parsed,
                            JSONLogLineOperation.EXECUTING_TASK,
                            line =>
                                !!(line.metadata as { task?: Task })?.task &&
                                line.metadata.task.Repository === task.Repository &&
                                line.metadata.task.Workspace === task.Workspace
                        )
                        const out = getCombinedLog(
                            parsed,
                            JSONLogLineOperation.EXECUTING_TASK,
                            line =>
                                !!(line.metadata as { task?: Task })?.task &&
                                line.metadata.task.Repository === task.Repository &&
                                line.metadata.task.Workspace === task.Workspace
                        )
                        return (
                            <li key={task.Repository + task.Workspace} className="list-group-item">
                                <h3>
                                    {!tuple && <TimerSandIcon className="icon-inline" />}
                                    {tuple && tuple[1] === undefined && <LoadingSpinner className="icon-inline" />}
                                    {tuple?.[1] !== undefined && <CheckIcon className="icon-inline" />}
                                    {task.Repository}
                                </h3>
                                <p>Running in {task.Workspace || '/'}</p>
                                <h4>Steps</h4>
                                <ul className="text-monospace mb-3">
                                    {task.Steps.map((step, index) => (
                                        <li key={index}>
                                            <p>
                                                {step.container}: {step.run}
                                            </p>
                                        </li>
                                    ))}
                                </ul>
                                <LogOutput text={out} className="mb-2" />

                                {task.CachedStepResultsFound && <p className="text-success">Cached result found!</p>}
                            </li>
                        )
                    })}
                </ul>
            )
            if (step) {
                result.push(step)
            }
            continue
        }
        if (operation === JSONLogLineOperation.EXECUTING_TASK) {
            continue
        }
        const step = transformLogsForOperationToStage(parsed, operation as JSONLogLineOperation, expand, now)
        if (step) {
            result.push(step)
        }
    }

    return result
}

function transformLogsForOperationToStage(
    logLines: JSONLogLine[],
    operation: JSONLogLineOperation,
    expand: boolean,
    now?: () => Date,
    executionLogEntryChildren?: React.ReactNode
): TimelineStage | undefined {
    const possibleTuple = findLogLineTuple(logLines, operation)
    if (!possibleTuple) {
        return undefined
    }
    const [start, end] = possibleTuple
    const out = getCombinedLog(logLines, operation)
    return {
        text: prettyOperationNames[operation],
        details: (
            <ExecutionLogEntry
                key={operation}
                logEntry={{
                    command: ['src', 'batch', 'preview', operation],
                    durationMilliseconds:
                        end && start ? parseISO(end.timestamp).getTime() - parseISO(start.timestamp).getTime() : null,
                    exitCode: end ? 0 : null,
                    key: operation,
                    out,
                    startTime: start.timestamp,
                }}
                now={now}
            >
                {executionLogEntryChildren}
            </ExecutionLogEntry>
        ),
        ...genericStage({ exitCode: end ? 0 : null, startTime: start.timestamp }, expand),
    }
}

function getCombinedLog(
    logLines: JSONLogLine[],
    operation: JSONLogLineOperation,
    filter?: (line: JSONLogLine) => boolean
): string {
    const allEntries = findLogLines(logLines, operation, undefined, filter)
    const out = allEntries
        .filter(entry => entry.message !== undefined && entry.message !== '')
        .map(entry => entry.message)
        .join('\n')
    return out
}

const teardownStage = (
    execution: BatchSpecExecutionFields,
    expand: boolean,
    now?: () => Date
): TimelineStage | undefined =>
    execution.steps.teardown.length === 0
        ? undefined
        : {
              text: 'Teardown',
              details: execution.steps.teardown.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(execution.steps.teardown, expand),
          }

const genericStage = <E extends { startTime: string; exitCode: number | null }>(
    value: E | E[],
    expand: boolean
): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expanded'> => {
    const finished = isArray(value) ? value.every(logEntry => logEntry.exitCode !== null) : value.exitCode !== null
    const success = isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: !finished ? <ProgressClockIcon /> : success ? <CheckIcon /> : <ErrorIcon />,
        date: isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expanded: expand || !(success || !finished),
    }
}

enum JSONLogLineOperation {
    PARSING_BATCH_SPEC = 'PARSING_BATCH_SPEC',
    RESOLVING_NAMESPACE = 'RESOLVING_NAMESPACE',
    PREPARING_DOCKER_IMAGES = 'PREPARING_DOCKER_IMAGES',
    DETERMINING_WORKSPACE_TYPE = 'DETERMINING_WORKSPACE_TYPE',
    RESOLVING_REPOSITORIES = 'RESOLVING_REPOSITORIES',
    DETERMINING_WORKSPACES = 'DETERMINING_WORKSPACES',
    CHECKING_CACHE = 'CHECKING_CACHE',
    EXECUTING_TASKS = 'EXECUTING_TASKS',
    EXECUTING_TASK = 'EXECUTING_TASK',
    UPLOADING_CHANGESET_SPECS = 'UPLOADING_CHANGESET_SPECS',
    CREATING_BATCH_SPEC = 'CREATING_BATCH_SPEC',
}

const prettyOperationNames: Record<JSONLogLineOperation, string> = {
    PARSING_BATCH_SPEC: 'Parsing batch spec',
    RESOLVING_NAMESPACE: 'Resolving namespace',
    PREPARING_DOCKER_IMAGES: 'Preparing docker images',
    DETERMINING_WORKSPACE_TYPE: 'Determining workspace type',
    RESOLVING_REPOSITORIES: 'Resolving repositories',
    DETERMINING_WORKSPACES: 'Determining workspaces',
    CHECKING_CACHE: 'Checking cache',
    EXECUTING_TASKS: 'Executing tasks',
    EXECUTING_TASK: 'Executing task',
    UPLOADING_CHANGESET_SPECS: 'Uploading changeset specs',
    CREATING_BATCH_SPEC: 'Creating batch spec',
}

enum JSONLogLineStatus {
    STARTED = 'STARTED',
    PROGRESS = 'PROGRESS',
    SUCCESS = 'SUCCESS',
    FAILED = 'FAILED',
}

interface ExecutingTaskJSONLogLine {
    operation: JSONLogLineOperation.EXECUTING_TASK
    timestamp: string
    status: JSONLogLineStatus
    message?: string
    metadata: {
        task: Task
    }
}

interface ExecutingTasksJSONLogLine {
    operation: JSONLogLineOperation.EXECUTING_TASKS
    timestamp: string
    status: JSONLogLineStatus
    message?: string
    metadata: {
        tasks: Task[] | null
    }
}

type JSONLogLine =
    | {
          operation: JSONLogLineOperation
          timestamp: string
          status: JSONLogLineStatus
          message?: string
          metadata?: any
      }
    | ExecutingTaskJSONLogLine
    | ExecutingTasksJSONLogLine

interface Step {
    run: string
    container: string
}

interface Task {
    Repository: string
    Workspace: string
    Steps: Step[]
    CachedStepResultsFound: boolean
}

function parseLogLines(out: string): JSONLogLine[] {
    return out
        .split('\n')
        .map(line => line.replace(/^std(out|err): /, ''))
        .map(line => {
            try {
                return JSON.parse(line) as JSONLogLine
            } catch (error) {
                return String(error)
            }
        })
        .filter((line): line is JSONLogLine => typeof line !== 'string')
}

function findLogLine(
    lines: JSONLogLine[],
    operation: JSONLogLineOperation,
    status: JSONLogLineStatus | undefined,
    filter?: (line: JSONLogLine) => boolean
): JSONLogLine | undefined {
    return findLogLines(lines, operation, status, filter)?.[0]
}

function findLogLines(
    lines: JSONLogLine[],
    operation: JSONLogLineOperation,
    status: JSONLogLineStatus | undefined,
    filter?: (line: JSONLogLine) => boolean
): JSONLogLine[] {
    return lines.filter(
        line =>
            line.operation === operation &&
            (status === undefined || line.status === status) &&
            (filter ? filter(line) : true)
    )
}

function findLogLineTuple(
    lines: JSONLogLine[],
    operation: JSONLogLineOperation,
    filter?: (line: JSONLogLine) => boolean
): [JSONLogLine] | [JSONLogLine, JSONLogLine] | undefined {
    const start = findLogLine(lines, operation, JSONLogLineStatus.STARTED, filter)
    if (!start) {
        return undefined
    }
    let end = findLogLine(lines, operation, JSONLogLineStatus.SUCCESS, filter)
    if (!end) {
        end = findLogLine(lines, operation, JSONLogLineStatus.FAILED, filter)
    }
    if (end) {
        return [start, end]
    }
    return [start]
}
