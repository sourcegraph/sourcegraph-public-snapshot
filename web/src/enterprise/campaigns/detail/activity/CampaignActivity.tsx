import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CommentList } from '../../../comments/CommentList'
import { CampaignImpactSummaryBar } from '../../common/CampaignImpactSummaryBar'
import { useCampaignImpactSummary } from '../../common/useCampaignImpactSummary'
import { CampaignBurndownChart } from '../burndownChart/CampaignBurndownChart'
import { CampaignTimeline } from '../timeline/CampaignTimeline'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'url'>

    className?: string
    history: H.History
}

/**
 * The activity related to a campaign.
 */
export const CampaignActivity: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const impactSummary = useCampaignImpactSummary(campaign)
    return (
        <div className={`campaign-activity ${className}`}>
            <CampaignImpactSummaryBar
                impactSummary={impactSummary}
                baseURL={campaign.url}
                urlFragmentOrPath="/"
                className="mb-4"
            />
            <CampaignBurndownChart {...props} campaign={campaign} className="mb-4" />
            <CampaignTimeline {...props} campaign={campaign} timelineItemsClassName="pb-6" />
            <CommentList {...props} commentable={campaign} />
        </div>
    )
}
