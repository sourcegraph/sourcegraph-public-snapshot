import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { Comment } from '../../comments/Comment'
import { CampaignAreaContext } from './CampaignArea'
import { CampaignHeaderEditableName } from './header/CampaignHeaderEditableName'

interface Props extends Pick<CampaignAreaContext, 'campaign' | 'onCampaignUpdate'>, ExtensionsControllerProps {
    className?: string

    history: H.History
}

/**
 * The overview for a single campaign.
 */
export const CampaignOverview: React.FunctionComponent<Props> = ({
    campaign,
    onCampaignUpdate,
    className = '',
    ...props
}) => (
    <div className={`campaign-overview ${className || ''}`}>
        <CampaignHeaderEditableName
            {...props}
            campaign={campaign}
            onCampaignUpdate={onCampaignUpdate}
            className="mb-3"
        />
        <Comment
            {...props}
            comment={campaign}
            onCommentUpdate={onCampaignUpdate}
            createdVerb="started campaign"
            emptyBody="No description provided."
            className="mb-3"
        />
    </div>
)
