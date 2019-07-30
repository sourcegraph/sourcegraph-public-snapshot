import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
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
        {campaign.description && (
            <Markdown dangerousInnerHTML={renderMarkdown(campaign.description)} className="mb-4" />
        )}
    </div>
)
