import PlusCircleIcon from 'mdi-react/PlusCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.ICreateThreadEvent

    className?: string
}

export const CreateThreadEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <TimelineItem icon={PlusCircleIcon} className={className} event={event}>
        <ActorLink actor={event.actor} /> created the {event.thread.__typename.toLowerCase()}{' '}
        <Link to={event.thread.url}>{event.thread.title}</Link>
    </TimelineItem>
)
