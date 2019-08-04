import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { Timestamp } from '../../../../components/time/Timestamp'
import { Timeline } from '../../../../components/timeline/Timeline'
import { CreateThreadEventTimelineItem } from './events/CreateThreadEventTimelineItem'
import { useCampaignTimelineItems } from './useCampaignTimelineItems'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
}

const LOADING = 'loading' as const

/**
 * A timeline of events related to the campaign.
 */
export const CampaignTimeline: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
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
                <Timeline tag="ol">
                    {timelineItems.nodes.map(event => {
                        const C = timelineItemComponentForEvent(event.__typename)
                        return <C key={event.id} event={event} className="campaign-timeline__item" />
                    })}
                </Timeline>
            )}
        </div>
    )
}

function timelineItemComponentForEvent(
    event: GQL.Event['__typename']
): React.ComponentType<{
    event: any // TODO!(sqs)
    className?: string
}> {
    switch (event) {
        case 'AddThreadToCampaignEvent':
            return CreateThreadEventTimelineItem
        case 'CreateThreadEvent':
            return CreateThreadEventTimelineItem // TODO!(sqs)
        case 'RemoveThreadFromCampaignEvent':
            return CreateThreadEventTimelineItem // TODO!(sqs)
        default:
            throw new Error(`unrecognized event type ${event}`)
    }
}
