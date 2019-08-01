import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { Timestamp } from '../../../components/time/Timestamp'
import { PersonLink } from '../../../user/PersonLink'
import { Comment } from '../../comments/Comment'
import { CampaignAreaContext } from './CampaignArea'
import { CampaignHeaderEditableName } from './header/CampaignHeaderEditableName'

interface Props
    extends Pick<CampaignAreaContext, 'campaign' | 'onCampaignUpdate'>,
        ExtensionsControllerNotificationProps {
    className?: string
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
        <div className="d-flex align-items-center py-3"></div>
        <hr className="my-3" />
        <Comment comment={campaign} createdVerb="created campaign" />
    </div>
)
