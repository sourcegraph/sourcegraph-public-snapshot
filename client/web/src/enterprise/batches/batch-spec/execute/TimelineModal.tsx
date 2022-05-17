import React, { useMemo } from 'react'

import VisuallyHidden from '@reach/visually-hidden'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { isDefined } from '@sourcegraph/common'
import { Button, Modal, Icon, Typography } from '@sourcegraph/wildcard'

import { ExecutionLogEntry } from '../../../../components/ExecutionLogEntry'
import { Timeline, TimelineStage } from '../../../../components/Timeline'
import { BatchSpecWorkspaceState, VisibleBatchSpecWorkspaceFields } from '../../../../graphql-operations'
import { ExecutorNode } from '../../../executors/ExecutorsListPage'

import styles from './TimelineModal.module.scss'

export interface TimelineModalProps {
    node: VisibleBatchSpecWorkspaceFields
    onCancel: () => void
}

export const TimelineModal: React.FunctionComponent<React.PropsWithChildren<TimelineModalProps>> = ({
    node,
    onCancel,
}) => (
    <Modal className={styles.modalBody} onDismiss={onCancel} aria-label="Execution timeline">
        <div className="d-flex justify-content-between">
            <Typography.H3 className="mb-0">Execution timeline</Typography.H3>
            <Button className="p-0 ml-2" onClick={onCancel} variant="icon">
                <VisuallyHidden>Close</VisuallyHidden>
                <Icon as={CloseIcon} />
            </Button>
        </div>
        <ExecutionTimeline node={node} />
        {node.executor && (
            <>
                <Typography.H4 className="mt-2">Executor</Typography.H4>
                <ExecutorNode node={node.executor} />
            </>
        )}
    </Modal>
)

interface ExecutionTimelineProps {
    node: VisibleBatchSpecWorkspaceFields
    className?: string

    /** For testing only. */
    now?: () => Date
    expandStage?: string
}

const ExecutionTimeline: React.FunctionComponent<React.PropsWithChildren<ExecutionTimelineProps>> = ({
    node,
    className,
    now,
    expandStage,
}) => {
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
    execution: VisibleBatchSpecWorkspaceFields,
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
    execution: VisibleBatchSpecWorkspaceFields,
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
    execution: VisibleBatchSpecWorkspaceFields,
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
    const finished = Array.isArray(value)
        ? value.every(logEntry => logEntry.exitCode !== null)
        : value.exitCode !== null
    const success = Array.isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: !finished ? <ProgressClockIcon /> : success ? <CheckIcon /> : <AlertCircleIcon />,
        date: Array.isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expanded: expand || !(success || !finished),
    }
}
