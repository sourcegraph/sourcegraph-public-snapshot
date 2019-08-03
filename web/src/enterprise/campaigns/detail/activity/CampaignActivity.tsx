import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CommentList } from '../../../comments/CommentList'
import { CampaignBurndownChart } from './CampaignBurndownChart'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
    history: H.History
}

/**
 * The activity related to a campaign.
 */
export const CampaignActivity: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => (
    <div className={`campaign-activity ${className}`}>
        <CampaignBurndownChart {...props} campaign={campaign} className="mb-4" />
        <CommentList {...props} commentable={campaign} />
    </div>
)
