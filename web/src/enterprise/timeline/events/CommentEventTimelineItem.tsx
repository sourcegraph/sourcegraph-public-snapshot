import { truncate } from 'lodash'
import MessageReplyIcon from 'mdi-react/MessageReplyIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.ICommentEvent

    className?: string
}

export const CommentEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => {
    if (event.comment.__typename !== 'CommentReply') {
        return null
    }
    if (!event.comment.parent || event.comment.parent.__typename !== 'Thread') {
        return null
    }
    return (
        <TimelineItem icon={MessageReplyIcon} className={className} event={event}>
            <ActorLink actor={event.actor} /> commented on{' '}
            <Link to={event.comment.parent.url}>{event.comment.parent.title}</Link>:{' '}
            {truncate(event.comment.bodyText, { separator: ' ', length: 50 })}
        </TimelineItem>
    )
}
