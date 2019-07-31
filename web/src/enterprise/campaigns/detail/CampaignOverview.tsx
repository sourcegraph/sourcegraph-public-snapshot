import React from 'react'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { Timestamp } from '../../../components/time/Timestamp'
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
        <div className="d-flex align-items-center py-3">
            <ThreadStatusBadge campaign={campaign} className="mr-3" />
            <div>
                <small>
                    Opened <Timestamp date={campaign.createdAt} /> by{' '}
                    <strong>
                        <PersonLink user={campaign.author} />
                    </strong>
                </small>
                {campaign.type === GQL.ThreadType.ISSUE && (
                    <ThreadStatusItemsProgressBar className="mt-1 mb-3" height="0.3rem" />
                )}
            </div>
            <ThreadStatusButton
                {...props}
                campaign={campaign}
                onThreadUpdate={onThreadUpdate}
                className="ml-2"
                buttonClassName="btn-link btn-sm"
            />
        </div>
        <hr className="my-0" />
        {campaign.description && (
            <Markdown dangerousInnerHTML={renderMarkdown(campaign.description)} className="mb-4" />
        )}
    </div>
)
