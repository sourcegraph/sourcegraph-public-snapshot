import DotsHorizontalCircleIcon from 'mdi-react/DotsHorizontalCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.IRequestReviewEvent

    className?: string
}

export const ReopenThreadEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <TimelineItem icon={DotsHorizontalCircleIcon} className={className} event={event}>
        <ActorLink actor={event.actor} /> reopened <Link to={event.thread.url}>{event.thread.title}</Link>
    </TimelineItem>
)
