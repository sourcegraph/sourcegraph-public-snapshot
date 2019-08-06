import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timeline } from '../../components/timeline/Timeline'
import { AddThreadToCampaignEventTimelineItem } from './events/AddThreadToCampaignEventTimelineItem'
import { CloseThreadEventTimelineItem } from './events/CloseThreadEventTimelineItem'
import { CreateThreadEventTimelineItem } from './events/CreateThreadEventTimelineItem'
import { MergeChangesetEventTimelineItem } from './events/MergeChangesetEventTimelineItem'
import { RemoveThreadFromCampaignEventTimelineItem } from './events/RemoveThreadFromCampaignEventTimelineItem'
import { ReopenThreadEventTimelineItem } from './events/ReopenThreadEventTimelineItem'
import { RequestReviewEventTimelineItem } from './events/RequestReviewEventTimelineItem'
import { ReviewEventTimelineItem } from './events/ReviewEventTimelineItem'

interface Props {
    events: Pick<GQL.IEventConnection, 'nodes'>

    className?: string
}

/**
 * A timeline showing timeline items.
 */
export const TimelineItems: React.FunctionComponent<Props> = ({ events, className }) => (
    <Timeline tag="ol" className={className}>
        {events.nodes.map((event, i) => {
            if (!event.id) {
                return (
                    <li key={i} className="border border-danger p-2">
                        Unrecognized timeline item: {event.__typename}
                    </li>
                )
            }
            const C = timelineItemComponentForEvent(event.__typename)
            return C ? <C key={event.id} event={event} className="timeline-items__item" /> : null
        })}
    </Timeline>
)

function timelineItemComponentForEvent(
    event: GQL.Event['__typename']
): React.ComponentType<{
    event: any // TODO!(sqs)
    className?: string
}> | null {
    switch (event) {
        case 'AddThreadToCampaignEvent':
            return AddThreadToCampaignEventTimelineItem
        case 'CreateThreadEvent':
            return CreateThreadEventTimelineItem
        case 'RemoveThreadFromCampaignEvent':
            return RemoveThreadFromCampaignEventTimelineItem
        case 'ReviewEvent':
            return ReviewEventTimelineItem // TODO!(sqs)
        case 'RequestReviewEvent':
            return RequestReviewEventTimelineItem // TODO!(sqs)
        case 'MergeChangesetEvent':
            return MergeChangesetEventTimelineItem
        case 'CloseThreadEvent':
            return CloseThreadEventTimelineItem
        case 'ReopenThreadEvent':
            return ReopenThreadEventTimelineItem
        case 'CommentOnThreadEvent':
            return CommentOnThreadEventTimelineItem
        default:
            return null
    }
}
