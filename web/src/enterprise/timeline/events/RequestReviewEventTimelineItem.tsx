import RecordIcon from 'mdi-react/RecordIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.IRequestReviewEvent

    className?: string
}

export const RequestReviewEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <TimelineItem icon={RecordIcon} className={className} event={event}>
        <ActorLink actor={event.actor} /> requested a review on <Link to={event.thread.url}>{event.thread.title}</Link>
    </TimelineItem>
)
