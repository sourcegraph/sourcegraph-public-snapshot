import React from 'react'
import ProgressCheckIcon from 'mdi-react/ProgressCheckIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import classNames from 'classnames'
import { CampaignFields } from '../../../graphql-operations'

interface CampaignStatsCardProps extends Pick<CampaignFields['changesets'], 'stats'> {}

export const CampaignStatsCard: React.FunctionComponent<CampaignStatsCardProps> = ({ stats }) => {
    const percentDone = stats.total === 0 ? 0 : (((stats.closed + stats.merged) / stats.total) * 100).toFixed(0)
    const isDone = stats.closed + stats.merged === stats.total
    let CampaignStatusIcon = ProgressCheckIcon
    if (isDone) {
        CampaignStatusIcon = CheckCircleOutlineIcon
    }
    return (
        <div className="card mt-2">
            <div className="card-body">
                <div className="d-flex justify-content-between align-items-center">
                    <div className="d-flex align-items-center">
                        <h1 className="d-inline mb-0">
                            <CampaignStatusIcon
                                className={classNames(
                                    'icon-inline mr-2',
                                    isDone && 'text-success',
                                    !isDone && 'text-muted'
                                )}
                            />
                        </h1>{' '}
                        {percentDone}% complete
                    </div>
                    <div className="text-muted">{stats.total} changesets total</div>
                    <div className="d-flex align-items-center">
                        <SourceBranchIcon className="icon-inline text-muted mr-2" /> {stats.unpublished} unpublished
                    </div>
                    <div className="d-flex align-items-center">
                        <SourceBranchIcon className="icon-inline text-success mr-2" /> {stats.open} open
                    </div>
                    <div className="d-flex align-items-center">
                        <SourceMergeIcon className="icon-inline text-merged mr-2" /> {stats.merged} merged
                    </div>
                    <div className="d-flex align-items-center">
                        <SourceBranchIcon className="icon-inline text-danger mr-2" /> {stats.closed} closed
                    </div>
                </div>
            </div>
        </div>
    )
}
