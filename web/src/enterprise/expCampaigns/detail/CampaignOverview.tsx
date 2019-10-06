import H from 'history'
import React from 'react'
import { SUPPORT_CAMPAIGN_UPDATES, USE_CAMPAIGN_RULES } from '..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { Timeline } from '../../../components/timeline/Timeline'
import { Comment } from '../../comments/Comment'
import { IsDraftTimelineBox } from '../common/IsDraftTimelineBox'
import { PublishDraftCampaignButton } from '../common/PublishDraftCampaign'
import { RulesTimelineBox } from '../common/RulesTimelineBox'
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
        <Timeline className="align-items-stretch mb-4">
            <Comment
                {...props}
                comment={campaign}
                onCommentUpdate={onCampaignUpdate}
                createdVerb="started campaign"
                emptyBody="No description provided."
            />
            {USE_CAMPAIGN_RULES && SUPPORT_CAMPAIGN_UPDATES && <RulesTimelineBox ruleContainer={campaign} />}
            {USE_CAMPAIGN_RULES && campaign.isDraft && (
                <IsDraftTimelineBox
                    noun="campaign"
                    action={
                        <PublishDraftCampaignButton
                            {...props}
                            campaign={campaign}
                            onComplete={onCampaignUpdate}
                            buttonClassName="btn-secondary"
                        />
                    }
                />
            )}
        </Timeline>
    </div>
)
