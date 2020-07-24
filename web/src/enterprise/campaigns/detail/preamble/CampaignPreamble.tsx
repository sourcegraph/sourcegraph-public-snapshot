import React from 'react'
import { Timeline } from '../../../../components/timeline/Timeline'
import { CampaignDescription } from '../CampaignDescription'
import H from 'history'
import { MinimalCampaign } from '../CampaignArea'
import { CampaignProgressCard } from './CampaignProgressCard'
import { CampaignUpdatesCard } from './CampaignUpdatesCard'
import { CampaignStateBadge } from '../../common/CampaignStateBadge'
import { Timestamp } from '../../../../components/time/Timestamp'
import { Link } from 'react-router-dom'
import { CampaignProgressBadge } from './CampaignProgressBadge'

interface Props {
    campaign: MinimalCampaign
    history: H.History
}

export const CampaignPreamble: React.FunctionComponent<Props> = ({ campaign, history }) => (
    <>
        <header>
            <h1 className="mb-1">{campaign.name}</h1>
            <div className="d-flex align-items-center">
                <CampaignStateBadge campaign={campaign} className="mr-2" />
                <CampaignProgressBadge
                    changesetCounts={
                        /* TODO(sqs) */
                        campaign.changesetCountsOverTime.length > 0 &&
                        campaign.changesetCountsOverTime.some(e => e.total > 0)
                            ? campaign.changesetCountsOverTime[campaign.changesetCountsOverTime.length - 1]
                            : /* TODO(sqs) */
                              { total: 107, merged: 23, closed: 8, open: 67 }
                    }
                    className="text-muted"
                />
                <span className="d-none">
                    Opened <Timestamp date={campaign.createdAt} /> by{' '}
                    <Link to={campaign.author.url}>
                        <strong>{campaign.author.username}</strong>
                    </Link>
                </span>
                <div className="flex-1" />
                {campaign.viewerCanAdminister && (
                    <>
                        <Link to={`${campaign.url}/edit`} className="btn btn-secondary mr-2">
                            Edit
                        </Link>
                        <Link to={`${campaign.url}/close`} className="btn btn-secondary">
                            Close
                        </Link>
                    </>
                )}
            </div>
        </header>
        <Timeline className="mt-3">
            <CampaignDescription campaign={campaign} history={history} className="w-100" />
            <CampaignProgressCard
                campaign={campaign}
                changesetCounts={
                    /* TODO(sqs) */
                    campaign.changesetCountsOverTime.length > 0 &&
                    campaign.changesetCountsOverTime.some(counts => counts.total > 0)
                        ? {
                              ...campaign.changesetCountsOverTime[campaign.changesetCountsOverTime.length - 1],
                              unpublished: 123 /* TODO(sqs) */,
                          }
                        : /* TODO(sqs) */
                          { total: 107, merged: 23, closed: 8, open: 67, unpublished: 9 }
                }
                history={history}
                className="w-100 mt-3"
            />
            <CampaignUpdatesCard
                campaign={{
                    ...campaign,
                    /* TODO(sqs) */ patchesSetAt: campaign.updatedAt,
                    patchSetter: campaign.author,
                }}
                history={history}
                className="w-100 mt-3"
            />
        </Timeline>
    </>
)
