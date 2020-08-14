import React from 'react'
import ProgressCheckIcon from 'mdi-react/ProgressCheckIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import classNames from 'classnames'
import { CampaignFields } from '../../../graphql-operations'
import { CampaignStateBadge } from './CampaignStateBadge'
import {
    ChangesetStatusUnpublished,
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
} from './changesets/ChangesetStatusCell'

interface CampaignStatsCardProps extends Pick<CampaignFields['changesets'], 'stats'> {
    closedAt: CampaignFields['closedAt']
    className?: string
}

export const CampaignStatsCard: React.FunctionComponent<CampaignStatsCardProps> = ({ stats, closedAt, className }) => {
    const percentComplete = stats.total === 0 ? 0 : (((stats.closed + stats.merged) / stats.total) * 100).toFixed(0)
    const isCompleted = stats.closed + stats.merged === stats.total
    let CampaignStatusIcon = ProgressCheckIcon
    if (isCompleted) {
        CampaignStatusIcon = CheckCircleOutlineIcon
    }
    return (
        <div className={classNames('card', className)}>
            <div className="card-body p-3">
                <div className="d-flex flex-wrap justify-content-between align-items-center">
                    <div className="d-flex align-items-center flex-grow-1">
                        <h2 className="m-0 mr-3">
                            <CampaignStateBadge isClosed={!!closedAt} />
                        </h2>
                        <h1 className="d-inline mb-0">
                            <CampaignStatusIcon
                                className={classNames(
                                    'icon-inline mr-2',
                                    isCompleted && 'text-success',
                                    !isCompleted && 'text-muted'
                                )}
                            />
                        </h1>{' '}
                        <span className="lead">{percentComplete}% complete</span>
                    </div>
                    <CampaignStatsTotalAction count={stats.total} />
                    <ChangesetStatusUnpublished
                        label={<span className="text-muted">{stats.unpublished} unpublished</span>}
                        className="flex-grow-0 flex-shrink-0 mx-3"
                    />
                    <ChangesetStatusOpen
                        label={<span className="text-muted">{stats.open} open</span>}
                        className="flex-grow-0 flex-shrink-0 mx-3"
                    />
                    <ChangesetStatusClosed
                        label={<span className="text-muted">{stats.closed} closed</span>}
                        className="flex-grow-0 flex-shrink-0 mx-3"
                    />
                    <ChangesetStatusMerged
                        label={<span className="text-muted">{stats.merged} merged</span>}
                        className="flex-grow-0 flex-shrink-0 ml-3"
                    />
                </div>
            </div>
        </div>
    )
}

export const CampaignStatsTotalAction: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="m-0 mr-3 flex-grow-0 flex-shrink-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <span className="campaign-stats-card__changesets-pill">
            <span className="badge badge-pill badge-secondary">{count}</span>
        </span>
        <span className="text-muted">changesets</span>
    </div>
)
