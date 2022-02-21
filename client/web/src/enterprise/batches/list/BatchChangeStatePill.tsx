import classNames from 'classnames'
import { includes } from 'lodash'
import HistoryIcon from 'mdi-react/HistoryIcon'
import React, { useMemo } from 'react'

import { Badge } from '@sourcegraph/wildcard'

import { BatchChangeState, BatchSpecState, Scalars } from '../../../graphql-operations'

import styles from './BatchChangeStatePill.module.scss'

export interface BatchChangeStatePillProps {
    className?: string
    state: BatchChangeState
    latestExecutionState?: BatchSpecState
    currentSpecID: Scalars['ID']
    latestSpecID?: Scalars['ID']
}

export const BatchChangeStatePill: React.FunctionComponent<BatchChangeStatePillProps> = ({
    className,
    state,
    latestExecutionState,
    currentSpecID,
    latestSpecID,
}) => {
    const isCurrentLatest = useMemo(() => currentSpecID === latestSpecID, [currentSpecID, latestSpecID])

    const hasExecutionVisible = useMemo(() => {
        // If the batch change is closed, we don't show any execution state.
        if (state === BatchChangeState.CLOSED) {
            return false
        }
        // If the batch change is a draft or open, we show execution state that the user
        // would want to take action on.
        if (includes([BatchSpecState.FAILED, BatchSpecState.PROCESSING, BatchSpecState.QUEUED], latestExecutionState)) {
            return true
        }
        // If the latest execution is complete, we show execution state if it has not yet been applied.
        return !isCurrentLatest && latestExecutionState === BatchSpecState.COMPLETED
    }, [latestExecutionState, state, isCurrentLatest])

    return (
        <div
            role="group"
            className={classNames(styles.pillGroup, className, {
                [styles.open]: state === BatchChangeState.OPEN,
                [styles.draft]: state === BatchChangeState.DRAFT,
                [styles.closed]: state === BatchChangeState.CLOSED,
            })}
        >
            <StatePill state={state} />
            {hasExecutionVisible && <ExecutionStatePill latestExecutionState={latestExecutionState} />}
        </div>
    )
}

const StatePill: React.FunctionComponent<Pick<BatchChangeStatePillProps, 'state'>> = ({ state }) => {
    switch (state) {
        case BatchChangeState.OPEN:
            return (
                <Badge variant="success" className={styles.statePill}>
                    Open
                </Badge>
            )
        case BatchChangeState.CLOSED:
            return (
                <Badge variant="danger" className={styles.statePill}>
                    Closed
                </Badge>
            )
        case BatchChangeState.DRAFT:
        default:
            return (
                <Badge variant="secondary" className={styles.statePill}>
                    Draft
                </Badge>
            )
    }
}

const ExecutionStatePill: React.FunctionComponent<Pick<BatchChangeStatePillProps, 'latestExecutionState'>> = ({
    latestExecutionState,
}) => {
    switch (latestExecutionState) {
        case BatchSpecState.PROCESSING:
        case BatchSpecState.QUEUED:
            return (
                <Badge
                    variant="warning"
                    className={styles.executionPill}
                    data-tooltip={`This batch change has a new spec ${
                        latestExecutionState === BatchSpecState.QUEUED
                            ? 'queued for execution'
                            : 'in the process of executing'
                    }.`}
                >
                    <HistoryIcon className={styles.executionIcon} />
                </Badge>
            )

        case BatchSpecState.COMPLETED:
            return (
                <Badge
                    variant="primary"
                    className={styles.executionPill}
                    data-tooltip="This batch change has a newer batch spec execution that is ready to be applied."
                >
                    <HistoryIcon className={styles.executionIcon} />
                </Badge>
            )
        case BatchSpecState.FAILED:
        default:
            return (
                <Badge
                    variant="danger"
                    className={styles.executionPill}
                    data-tooltip="The latest batch spec execution for this batch change failed."
                >
                    <HistoryIcon className={styles.executionIcon} />
                </Badge>
            )
    }
}
