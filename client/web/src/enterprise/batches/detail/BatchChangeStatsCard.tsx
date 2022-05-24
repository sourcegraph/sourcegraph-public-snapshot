import React from 'react'

import classNames from 'classnames'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import ProgressCheckIcon from 'mdi-react/ProgressCheckIcon'

import { pluralize } from '@sourcegraph/common'
import { Badge, Icon, Typography } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../components/diff/DiffStat'
import { BatchChangeFields } from '../../../graphql-operations'
import { BatchChangeStatePill } from '../list/BatchChangeStatePill'

import {
    ChangesetStatusUnpublished,
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
    ChangesetStatusDraft,
    ChangesetStatusArchived,
} from './changesets/ChangesetStatusCell'

import styles from './BatchChangeStatsCard.module.scss'

interface BatchChangeStatsCardProps {
    batchChange: Pick<BatchChangeFields, 'diffStat' | 'changesetsStats' | 'state'>
    className?: string
}

// Rounds percent down to the nearest integer (you don't say 1/50/100% complete until at
// least 1/50/100% is actually completed).
const formatDisplayPercent = (percent: number): string => `${Math.floor(percent)}%`

export const BatchChangeStatsCard: React.FunctionComponent<React.PropsWithChildren<BatchChangeStatsCardProps>> = ({
    batchChange,
    className,
}) => {
    const { changesetsStats: stats, diffStat } = batchChange
    const percentComplete = stats.total === 0 ? 0 : ((stats.closed + stats.merged + stats.deleted) / stats.total) * 100
    const isCompleted = stats.closed + stats.merged + stats.deleted === stats.total
    let BatchChangeStatusIcon = ProgressCheckIcon
    if (isCompleted) {
        BatchChangeStatusIcon = CheckCircleOutlineIcon
    }
    return (
        <div className={classNames(className)}>
            <div className="d-flex flex-wrap align-items-center flex-grow-1">
                <Typography.H2 className="m-0">
                    {/*
                        a11y-ignore
                        Rule: "color-contrast" (Elements must have sufficient color contrast)
                        GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                    */}
                    <BatchChangeStatePill
                        state={batchChange.state}
                        className={classNames('a11y-ignore', styles.batchChangeStatsCardStateBadge)}
                    />
                </Typography.H2>
                <div className={classNames(styles.batchChangeStatsCardDivider, 'mx-3')} />
                <div className="d-flex align-items-center">
                    <Typography.H1 className="d-inline mb-0" aria-label="Batch Change Status">
                        <Icon
                            className={classNames('mr-2', isCompleted ? 'text-success' : 'text-muted')}
                            as={BatchChangeStatusIcon}
                            aria-label="Batch Change Status Icon"
                        />
                    </Typography.H1>{' '}
                    <span className={classNames(styles.batchChangeStatsCardCompleteness, 'lead text-nowrap')}>
                        {formatDisplayPercent(percentComplete)} complete
                    </span>
                </div>
                <div className={classNames(styles.batchChangeStatsCardDivider, 'd-none d-md-block mx-3')} />
                <DiffStatStack className={styles.batchChangeStatsCardDiffStat} {...diffStat} />
                <div className="d-flex flex-wrap justify-content-end flex-grow-1">
                    <BatchChangeStatsTotalAction count={stats.total} />
                    <ChangesetStatusUnpublished
                        label={<span className="text-muted">{stats.unpublished} Unpublished</span>}
                        className={classNames(styles.batchChangeStatsCardStat, 'd-flex flex-grow-0 px-2 text-truncate')}
                    />
                    <ChangesetStatusDraft
                        label={<span className="text-muted">{stats.draft} Draft</span>}
                        className={classNames(styles.batchChangeStatsCardStat, 'd-flex flex-grow-0 px-2 text-truncate')}
                    />
                    <ChangesetStatusOpen
                        label={<span className="text-muted">{stats.open} Open</span>}
                        className={classNames(styles.batchChangeStatsCardStat, 'd-flex flex-grow-0 px-2 text-truncate')}
                    />
                    <ChangesetStatusClosed
                        label={<span className="text-muted">{stats.closed} Closed</span>}
                        className={classNames(styles.batchChangeStatsCardStat, 'd-flex flex-grow-0 px-2 text-truncate')}
                    />
                    <ChangesetStatusMerged
                        label={<span className="text-muted">{stats.merged} Merged</span>}
                        className={classNames(styles.batchChangeStatsCardStat, 'd-flex flex-grow-0 pl-2 text-truncate')}
                    />
                    <ChangesetStatusArchived
                        label={<span className="text-muted">{stats.archived} Archived</span>}
                        className={classNames(styles.batchChangeStatsCardStat, 'd-flex flex-grow-0 pl-2 text-truncate')}
                    />
                </div>
            </div>
        </div>
    )
}

export const BatchChangeStatsTotalAction: React.FunctionComponent<React.PropsWithChildren<{ count: number }>> = ({
    count,
}) => (
    <div
        className={classNames(
            styles.batchChangeStatsCardStat,
            'm-0 flex-grow-0 pr-2 text-truncate text-nowrap d-flex flex-column align-items-center justify-content-center'
        )}
    >
        <span className={styles.batchChangeStatsCardChangesetsPill}>
            <Badge variant="secondary" pill={true}>
                {count}
            </Badge>
        </span>
        <span className="text-muted">{pluralize('Changeset', count)}</span>
    </div>
)
