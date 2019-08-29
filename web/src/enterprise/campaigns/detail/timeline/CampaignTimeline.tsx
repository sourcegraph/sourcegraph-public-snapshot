import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { TimelineItems } from '../../../timeline/TimelineItems'
import { useCampaignTimelineItems } from './useCampaignTimelineItems'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
    timelineItemsClassName?: string
}

const LOADING = 'loading' as const

/**
 * A timeline of events related to the campaign.
 */
export const CampaignTimeline: React.FunctionComponent<Props> = ({
    campaign,
    className = '',
    timelineItemsClassName,
}) => {
    const [timelineItems] = useCampaignTimelineItems(campaign)
    return (
        <div className={`campaign-timeline ${className}`}>
            {timelineItems === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(timelineItems) ? (
                <div className="alert alert-danger">{timelineItems.message}</div>
            ) : timelineItems.totalCount === 0 ? (
                <span className="text-muted">No events.</span>
            ) : (
                <TimelineItems events={timelineItems} className={timelineItemsClassName} />
            )}
        </div>
    )
}
