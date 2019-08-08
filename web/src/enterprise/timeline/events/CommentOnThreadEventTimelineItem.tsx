import MessageReplyIcon from 'mdi-react/MessageReplyIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.ICommentOnThreadEvent

    className?: string
}

export const CommentOnThreadEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <TimelineItem icon={MessageReplyIcon} className={className} event={event}>
        <ActorLink actor={event.actor} /> commented on <Link to={event.thread.url}>{event.thread.title}</Link>
    </TimelineItem>
)
