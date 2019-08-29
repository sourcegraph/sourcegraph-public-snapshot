import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timeline } from '../../components/timeline/Timeline'
import { AddThreadToCampaignEventTimelineItem } from './events/AddThreadToCampaignEventTimelineItem'
import { CloseThreadEventTimelineItem } from './events/CloseThreadEventTimelineItem'
import { CommentEventTimelineItem } from './events/CommentEventTimelineItem'
import { CommentOnThreadEventTimelineItem } from './events/CommentOnThreadEventTimelineItem'
import { CreateThreadEventTimelineItem } from './events/CreateThreadEventTimelineItem'
import { MergeThreadEventTimelineItem } from './events/MergeThreadEventTimelineItem'
import { RemoveThreadFromCampaignEventTimelineItem } from './events/RemoveThreadFromCampaignEventTimelineItem'
import { ReopenThreadEventTimelineItem } from './events/ReopenThreadEventTimelineItem'
import { RequestReviewEventTimelineItem } from './events/RequestReviewEventTimelineItem'
import { ReviewEventTimelineItem } from './events/ReviewEventTimelineItem'
import { ThreadDiagnosticEdgeEventTimelineItem } from './events/ThreadDiagnosticEdgeEventTimelineItem'

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
            if (!event.__typename) {
                return null // occurs for event types with no `... on XyzEvent` in GraphQL query
            }
            if (!event.id) {
                return (
                    // eslint-disable-next-line react/no-array-index-key
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
        case 'CreateThreadEvent':
            return CreateThreadEventTimelineItem
        case 'CommentEvent':
            return CommentEventTimelineItem
        case 'AddThreadToCampaignEvent':
            return AddThreadToCampaignEventTimelineItem
        case 'RemoveThreadFromCampaignEvent':
            return RemoveThreadFromCampaignEventTimelineItem
        case 'ReviewEvent':
            return ReviewEventTimelineItem // TODO!(sqs)
        case 'RequestReviewEvent':
            return RequestReviewEventTimelineItem // TODO!(sqs)
        case 'MergeThreadEvent':
            return MergeThreadEventTimelineItem
        case 'CloseThreadEvent':
            return CloseThreadEventTimelineItem
        case 'ReopenThreadEvent':
            return ReopenThreadEventTimelineItem
        case 'CommentOnThreadEvent':
            // TODO!(sqs): remove this in favor of CommentEvent
            return CommentOnThreadEventTimelineItem
        case 'AddDiagnosticToThreadEvent':
            return ThreadDiagnosticEdgeEventTimelineItem
        case 'RemoveDiagnosticFromThreadEvent':
            return ThreadDiagnosticEdgeEventTimelineItem
        default:
            return null
    }
}
