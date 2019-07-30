import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { CampaignAreaContext } from './CampaignArea'

interface Props extends Pick<CampaignAreaContext, 'campaign'> {
    className?: string
}

/**
 * The overview for a single campaign.
 */
export const CampaignOverview: React.FunctionComponent<Props> = ({ campaign, className = '' }) => (
    <div className={`campaign-overview ${className || ''}`}>
        <h2>{campaign.name}</h2>
        {campaign.description && (
            <Markdown dangerousInnerHTML={renderMarkdown(campaign.description)} className="mb-4" />
        )}
    </div>
)
