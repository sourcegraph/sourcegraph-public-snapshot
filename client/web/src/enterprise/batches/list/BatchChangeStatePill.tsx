import React, { useMemo } from 'react'

import { mdiHistory } from '@mdi/js'
import classNames from 'classnames'

import { Badge, Icon } from '@sourcegraph/wildcard'

import { BatchChangeState, BatchSpecState, Scalars } from '../../../graphql-operations'

import styles from './BatchChangeStatePill.module.scss'

// A batch spec state is actionable if it's not pending, canceling, or canceled.
const actionableBatchSpecStates = [
    BatchSpecState.COMPLETED,
    BatchSpecState.FAILED,
    BatchSpecState.PROCESSING,
    BatchSpecState.QUEUED,
] as const
type ActionableBatchSpecState = typeof actionableBatchSpecStates[number]

const isLatestExecutionActionable = (executionState: BatchSpecState): executionState is ActionableBatchSpecState =>
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    actionableBatchSpecStates.includes(executionState as any)

export interface BatchChangeStatePillProps {
    className?: string
    state: BatchChangeState
    latestExecutionState?: BatchSpecState
    currentSpecID?: Scalars['ID']
    latestSpecID?: Scalars['ID']
}

export const BatchChangeStatePill: React.FunctionComponent<React.PropsWithChildren<BatchChangeStatePillProps>> = ({
    className,
    state,
    latestExecutionState,
    currentSpecID,
    latestSpecID,
}) => {
    const isCompleteAndApplied = useMemo(
        () => currentSpecID === latestSpecID && latestExecutionState === BatchSpecState.COMPLETED,
        [currentSpecID, latestSpecID, latestExecutionState]
    )

    // We only show the execution status part of the pill if:
    // - the batch change is not closed
    // - the latest execution state is actionable
    // - the latest execution is not already complete and applied
    const executionStatePill =
        latestExecutionState &&
        state !== BatchChangeState.CLOSED &&
        isLatestExecutionActionable(latestExecutionState) &&
        !isCompleteAndApplied ? (
            <ExecutionStatePill latestExecutionState={latestExecutionState} />
        ) : null

    return (
        <div
            role="group"
            className={classNames(styles.pillGroup, className, {
                [styles.open]: state === BatchChangeState.OPEN,
                [styles.draft]: state === BatchChangeState.DRAFT,
                [styles.closed]: state === BatchChangeState.CLOSED,
            })}
            aria-label={`${state} status`}
        >
            <StatePill state={state} />
            {executionStatePill}
        </div>
    )
}

const StatePill: React.FunctionComponent<React.PropsWithChildren<Pick<BatchChangeStatePillProps, 'state'>>> = ({
    state,
}) => {
    switch (state) {
        case BatchChangeState.OPEN:
            return (
                <Badge variant="success" className={styles.statePill} aria-hidden={true}>
                    Open
                </Badge>
            )
        case BatchChangeState.CLOSED:
            return (
                <Badge variant="danger" className={styles.statePill} aria-hidden={true}>
                    Closed
                </Badge>
            )
        case BatchChangeState.DRAFT:
        default:
            return (
                <Badge variant="secondary" className={styles.statePill} aria-hidden={true}>
                    Draft
                </Badge>
            )
    }
}

const ExecutionStatePill: React.FunctionComponent<
    React.PropsWithChildren<{ latestExecutionState: ActionableBatchSpecState }>
> = ({ latestExecutionState }) => {
    switch (latestExecutionState) {
        case BatchSpecState.PROCESSING:
        case BatchSpecState.QUEUED:
            return (
                <Badge
                    variant="warning"
                    tooltip={`This batch change has a new spec ${
                        latestExecutionState === BatchSpecState.QUEUED
                            ? 'queued for execution'
                            : 'in the process of executing'
                    }.`}
                    className={styles.executionPill}
                >
                    <Icon
                        className={styles.executionIcon}
                        svgPath={mdiHistory}
                        inline={false}
                        aria-label={`This batch change has a new spec ${
                            latestExecutionState === BatchSpecState.QUEUED
                                ? 'queued for execution'
                                : 'in the process of executing'
                        }.`}
                    />
                </Badge>
            )

        case BatchSpecState.COMPLETED:
            return (
                <Badge
                    variant="primary"
                    tooltip="This batch change has a newer batch spec execution that is ready to be applied."
                    className={styles.executionPill}
                >
                    <Icon className={styles.executionIcon} svgPath={mdiHistory} inline={false} aria-hidden={true} />
                </Badge>
            )
        case BatchSpecState.FAILED:
        default:
            return (
                <Badge
                    variant="danger"
                    tooltip="The latest batch spec execution for this batch change failed."
                    className={styles.executionPill}
                >
                    <Icon className={styles.executionIcon} svgPath={mdiHistory} inline={false} aria-hidden={true} />
                </Badge>
            )
    }
}
