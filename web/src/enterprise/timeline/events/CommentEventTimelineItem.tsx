import MessageReplyIcon from 'mdi-react/MessageReplyIcon'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.ICommentEvent

    className?: string
}

export const CommentEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => {
    return (
        <TimelineItem icon={MessageReplyIcon} className={className} event={event}>
            <ActorLink actor={event.actor} /> commented on TODO - TODO show comment excerpt
        </TimelineItem>
    )
}
