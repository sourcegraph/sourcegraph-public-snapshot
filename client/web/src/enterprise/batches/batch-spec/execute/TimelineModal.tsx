import React, { useMemo } from 'react'

import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { isDefined } from '@sourcegraph/common'
import { Button, Modal, Icon, H3, H4 } from '@sourcegraph/wildcard'

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
    <Modal className={styles.modalBody} position="center" onDismiss={onCancel} aria-label="Execution timeline">
        <div className={styles.modalHeader}>
            <H3 className="mb-0">Execution timeline</H3>
            <Button className="p-0 ml-2" onClick={onCancel} variant="icon">
                <VisuallyHidden>Close</VisuallyHidden>
                <Icon aria-hidden={true} as={CloseIcon} />
            </Button>
        </div>
        <div className={styles.modalContent}>
            <ExecutionTimeline node={node} />
            {node.executor && (
                <>
                    <H4 className="mt-2">Executor</H4>
                    <ExecutorNode node={node.executor} />
                </>
            )}
        </div>
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
            {
                icon: <Icon as={TimerSandIcon} aria-label="Success" />,
                text: 'Queued',
                date: node.queuedAt,
                className: 'bg-success',
            },
            {
                icon: <Icon as={CheckIcon} aria-label="Success" />,
                text: 'Began processing',
                date: node.startedAt,
                className: 'bg-success',
            },

            setupStage(node, expandStage === 'setup', now),
            batchPreviewStage(node, expandStage === 'srcPreview', now),
            teardownStage(node, expandStage === 'teardown', now),

            node.state === BatchSpecWorkspaceState.COMPLETED
                ? {
                      icon: <Icon as={CheckIcon} aria-label="Success" />,
                      text: 'Finished',
                      date: node.finishedAt,
                      className: 'bg-success',
                  }
                : node.state === BatchSpecWorkspaceState.CANCELED
                ? {
                      icon: <Icon as={AlertCircleIcon} aria-label="Success" />,
                      text: 'Canceled',
                      date: node.finishedAt,
                      className: 'bg-secondary',
                  }
                : {
                      icon: <Icon as={AlertCircleIcon} aria-label="Failed" />,
                      text: 'Failed',
                      date: node.finishedAt,
                      className: 'bg-danger',
                  },
        ],
        [expandStage, node, now]
    )
    return (
        <Timeline
            stages={stages.filter(isDefined)}
            now={now}
            className={classNames(className, styles.timelineMargin)}
        />
    )
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
        icon: !finished ? (
            <Icon as={ProgressClockIcon} aria-label="success" />
        ) : success ? (
            <Icon as={CheckIcon} aria-label="Success" />
        ) : (
            <Icon as={AlertCircleIcon} aria-label="Failed" />
        ),
        date: Array.isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expanded: expand || !(success || !finished),
    }
}
