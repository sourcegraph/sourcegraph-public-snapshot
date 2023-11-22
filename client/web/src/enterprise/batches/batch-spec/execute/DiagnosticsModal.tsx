import React, { useMemo } from 'react'

import { mdiClose, mdiTimerSand, mdiCheck, mdiAlertCircle, mdiProgressClock } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'

import { Button, Modal, Icon, H3, H4 } from '@sourcegraph/wildcard'

import { isDefined } from '../../../../codeintel/util/helpers'
import { ExecutionLogEntry } from '../../../../components/ExecutionLogEntry'
import { Timeline, type TimelineStage } from '../../../../components/Timeline'
import { BatchSpecWorkspaceState, type VisibleBatchSpecWorkspaceFields } from '../../../../graphql-operations'
import { ExecutorNode } from '../../../executors/instances/ExecutorNode'

import styles from './DiagnosticsModal.module.scss'

interface DiagnosticsModalProps {
    node: VisibleBatchSpecWorkspaceFields
    onCancel: () => void
}

export const DiagnosticsModal: React.FunctionComponent<React.PropsWithChildren<DiagnosticsModalProps>> = ({
    node,
    onCancel,
}) => (
    <Modal className={styles.modalBody} position="center" onDismiss={onCancel} aria-label="Execution timeline">
        <div className={styles.modalHeader}>
            <H3 className="mb-0">Execution timeline</H3>
            <Button className="p-0 ml-2" onClick={onCancel} variant="icon">
                <VisuallyHidden>Close</VisuallyHidden>
                <Icon aria-hidden={true} svgPath={mdiClose} />
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
    expandedStage?: string
}

const ExecutionTimeline: React.FunctionComponent<React.PropsWithChildren<ExecutionTimelineProps>> = ({
    node,
    className,
    now,
    expandedStage,
}) => {
    const stages = useMemo(
        () =>
            [
                {
                    icon: <Icon aria-label="Success" svgPath={mdiTimerSand} />,
                    text: 'Queued',
                    date: node.queuedAt,
                    className: 'bg-success',
                },
                {
                    icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                    text: 'Began processing',
                    date: node.startedAt,
                    className: 'bg-success',
                },

                setupStage(node, expandedStage === 'setup', now),
                batchPreviewStage(node, expandedStage === 'srcPreview', now),
                teardownStage(node, expandedStage === 'teardown', now),

                node.state === BatchSpecWorkspaceState.COMPLETED
                    ? {
                          icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                          text: 'Finished',
                          date: node.finishedAt,
                          className: 'bg-success',
                      }
                    : node.state === BatchSpecWorkspaceState.CANCELED
                    ? {
                          icon: <Icon aria-label="Success" svgPath={mdiAlertCircle} />,
                          text: 'Canceled',
                          date: node.finishedAt,
                          className: 'bg-secondary',
                      }
                    : {
                          icon: <Icon aria-label="Failed" svgPath={mdiAlertCircle} />,
                          text: 'Failed',
                          date: node.finishedAt,
                          className: 'bg-danger',
                      },
            ]
                .filter(isDefined)
                .filter<TimelineStage>((stage): stage is TimelineStage => stage.date !== null),
        [expandedStage, node, now]
    )
    return <Timeline stages={stages} now={now} className={className} />
}

const setupStage = (
    execution: VisibleBatchSpecWorkspaceFields,
    expandedByDefault: boolean,
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
              ...genericStage(execution.stages.setup, expandedByDefault),
          }
}

const batchPreviewStage = (
    execution: VisibleBatchSpecWorkspaceFields,
    expandedByDefault: boolean,
    now?: () => Date
): TimelineStage | undefined => {
    if (execution.stages === null) {
        return undefined
    }
    return execution.stages.srcExec.length === 0
        ? undefined
        : {
              text: 'Create batch spec preview',
              details: execution.stages.srcExec.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(execution.stages.srcExec, expandedByDefault),
          }
}

const teardownStage = (
    execution: VisibleBatchSpecWorkspaceFields,
    expandedByDefault: boolean,
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
              ...genericStage(execution.stages.teardown, expandedByDefault),
          }
}

const genericStage = <E extends { startTime: string; exitCode: number | null }>(
    value: E | E[],
    expandedByDefault: boolean
): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expandedByDefault'> => {
    const finished = Array.isArray(value)
        ? value.every(logEntry => logEntry.exitCode !== null)
        : value.exitCode !== null
    const success = Array.isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: !finished ? (
            <Icon aria-label="success" svgPath={mdiProgressClock} />
        ) : success ? (
            <Icon aria-label="Success" svgPath={mdiCheck} />
        ) : (
            <Icon aria-label="Failed" svgPath={mdiAlertCircle} />
        ),
        date: Array.isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expandedByDefault: expandedByDefault || !(success || !finished),
    }
}
