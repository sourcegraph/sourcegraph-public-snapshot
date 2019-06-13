import PlusCircleIcon from 'mdi-react/PlusCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.IAddThreadToCampaignEvent

    className?: string
}

export const AddThreadToCampaignEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <TimelineItem icon={PlusCircleIcon} className={className} event={event}>
        <ActorLink actor={event.actor} /> added the {event.thread.__typename.toLowerCase()}{' '}
        <Link to={event.thread.url}>{event.thread.title}</Link> to the campaign{' '}
        <Link to={event.campaign.url}>{event.campaign.name}</Link>
    </TimelineItem>
)
