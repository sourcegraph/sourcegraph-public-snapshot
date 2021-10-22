import { parseISO } from 'date-fns/esm'
import { isArray, isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
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
import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'
import { DiffStat } from '@sourcegraph/web/src/components/diff/DiffStat'
import { FileDiffConnection } from '@sourcegraph/web/src/components/diff/FileDiffConnection'
import { FileDiffNode } from '@sourcegraph/web/src/components/diff/FileDiffNode'
import { ExecutionLogEntry } from '@sourcegraph/web/src/components/ExecutionLogEntry'
import { FilteredConnectionQueryArguments } from '@sourcegraph/web/src/components/FilteredConnection'
import { Timeline, TimelineStage } from '@sourcegraph/web/src/components/Timeline'
import { Container, PageHeader, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { BatchSpecExecutionFields, Scalars } from '../../../graphql-operations'
import { BatchSpec } from '../BatchSpec'

import {
    cancelBatchSpecExecution,
    fetchBatchSpecExecution as _fetchBatchSpecExecution,
    queryBatchSpecWorkspaceStepFileDiffs,
} from './backend'

export interface BatchSpecExecutionDetailsPageProps extends ThemeProps {
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
    isLightTheme,
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
                                isLightTheme={isLightTheme}
                            />
                        )}
                    </Container>
                </div>
            </div>

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

const WorkspaceNode: React.FunctionComponent<
    {
        node: Workspace
    } & ThemeProps
> = ({ node, isLightTheme }) => {
    const a = ''
    return (
        <>
            <div className="d-flex justify-content-between">
                <h4>
                    <WorkspaceStateIcon node={node} /> {node.repository.name}
                </h4>
                {node.startedAt && <Duration start={node.startedAt} end={node.finishedAt ?? undefined} />}
            </div>
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
                    <Tabs size="medium">
                        <TabList>
                            <Tab key="logs">Logs</Tab>
                            <Tab key="output-variables">Output variables</Tab>
                            <Tab key="diff">Diff</Tab>
                            <Tab key="files-env">Files / Env</Tab>
                            <Tab key="command-container">Commands / container</Tab>
                            <Tab key="timeline">Timeline</Tab>
                        </TabList>
                        <TabPanels>
                            <TabPanel key="logs">
                                <pre className="card p-2">{step.outputLines?.join('\n')}</pre>
                            </TabPanel>
                            <TabPanel key="output-variables">
                                <ul>
                                    {step.outputVariables?.map(variable => (
                                        <li key={variable.name}>
                                            {variable.name}: {variable.value}
                                        </li>
                                    ))}
                                </ul>
                            </TabPanel>
                            <TabPanel key="diff">
                                <WorkspaceStepFileDiffConnection
                                    isLightTheme={isLightTheme}
                                    step={index + 1}
                                    workspace={node}
                                />
                            </TabPanel>
                            <TabPanel key="files-env">
                                <ul>
                                    {step.environment.map(variable => (
                                        <li key={variable.name}>
                                            {variable.name}: {variable.value}
                                        </li>
                                    ))}
                                </ul>
                            </TabPanel>
                            <TabPanel key="command-container">
                                <p className="text-monospace">{step.run}</p>
                                <p className="text-monospace mb-0">{step.container}</p>
                            </TabPanel>
                            <TabPanel key="timeline">
                                <ExecutionTimeline node={node} />
                            </TabPanel>
                        </TabPanels>
                    </Tabs>
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

interface ExecutionTimelineProps {
    node: Workspace
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
                ? { icon: <ErrorIcon />, text: 'Canceled', date: node.finishedAt, className: 'bg-secondary' }
                : { icon: <ErrorIcon />, text: 'Failed', date: node.finishedAt, className: 'bg-danger' },
        ],
        [expandStage, node, now]
    )
    return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
}

const setupStage = (execution: Workspace, expand: boolean, now?: () => Date): TimelineStage | undefined => {
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

const batchPreviewStage = (execution: Workspace, expand: boolean, now?: () => Date): TimelineStage | undefined => {
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

const teardownStage = (execution: Workspace, expand: boolean, now?: () => Date): TimelineStage | undefined => {
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
        icon: !finished ? <ProgressClockIcon /> : success ? <CheckIcon /> : <ErrorIcon />,
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
        workspace: Workspace
        step: number
    } & ThemeProps
> = ({ workspace, step, isLightTheme }) => {
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryBatchSpecWorkspaceStepFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                node: workspace.id,
                step,
            }),
        [workspace.id, step]
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
