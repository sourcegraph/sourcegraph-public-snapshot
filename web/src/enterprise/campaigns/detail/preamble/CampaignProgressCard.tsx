import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import ProgressCheckIcon from 'mdi-react/ProgressCheckIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import H from 'history'
import { Link } from 'react-router-dom'
import { CampaignChangesetsEditButton } from '../changesets/CampaignChangesetsEditButton'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { CampaignDiffStat } from '../CampaignDiffStat'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id' | 'url' | 'diffStat' | 'viewerCanAdminister'>

    /** The latest changeset states. */
    changesetCounts: Pick<GQL.IChangesetCounts, 'total' | 'merged' | 'closed' | 'open'> & {
        unpublished: number /* TODO(sqs): blocked on unification of changesets and patches */
    }

    history: H.History
    className?: string
}

/**
 * A summary of the campaign's progress toward completion, shown in the campaign preamble
 * "timeline".
 */
export const CampaignProgressCard: React.FunctionComponent<Props> = ({ campaign, changesetCounts, className = '' }) => {
    const completed = changesetCounts.merged + changesetCounts.closed
    const iconClassName = 'h3 mb-0 mr-2 icon-inline'
    return (
        <div className={`card ${className}`}>
            <div className="card-body d-flex align-items-center flex-wrap">
                {changesetCounts.total > 0 && completed === changesetCounts.total ? (
                    <CheckCircleOutlineIcon className={`${iconClassName} text-success`} />
                ) : (
                    <ProgressCheckIcon className={`${iconClassName} text-muted`} />
                )}
                {changesetCounts.total === 0 ? (
                    <strong>No changesets.</strong>
                ) : (
                    <>
                        <strong className="mr-3">
                            {Math.floor((completed / changesetCounts.total) * 100)}% complete
                        </strong>
                        <span className="text-muted mr-2">
                            {changesetCounts.total} {pluralize('changeset', changesetCounts.total)} total
                        </span>
                        <div className="d-flex align-items-center flex-1">
                            <span className="p-2 mr-2 text-nowrap">
                                <SourcePullIcon className="icon-inline text-muted" /> {changesetCounts.unpublished}{' '}
                                unpublished
                            </span>
                            <span className="p-2 mr-2 text-nowrap">
                                <SourcePullIcon className="icon-inline text-success" /> {changesetCounts.open} open
                            </span>
                            <span className="p-2 mr-2 text-nowrap">
                                <SourceMergeIcon className="icon-inline text-merged" /> {changesetCounts.merged} merged
                            </span>
                            <span className="p-2 mr-2 text-nowrap">
                                <SourcePullIcon className="icon-inline text-danger" /> {changesetCounts.closed} closed
                            </span>
                            <div className="flex-1" />
                            <CampaignDiffStat campaign={campaign} className="ml-2 mr-2 mb-0" />
                        </div>
                    </>
                )}
            </div>
            {campaign.viewerCanAdminister && (
                <footer className="card-footer small text-muted">
                    To add a changeset to this campaign,{' '}
                    <CampaignChangesetsEditButton
                        campaign={campaign}
                        buttonClassName="font-weight-bold btn btn-sm btn-link p-0"
                    >
                        update the campaign plan
                    </CampaignChangesetsEditButton>
                    .
                </footer>
            )}
        </div>
    )
}
