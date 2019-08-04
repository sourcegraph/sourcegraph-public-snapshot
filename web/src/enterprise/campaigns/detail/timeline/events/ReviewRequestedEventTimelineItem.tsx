import PlusCircleIcon from 'mdi-react/PlusCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { PersonLink } from '../../../../../user/PersonLink'
import { CampaignTimelineItem } from '../CampaignTimelineItem'

interface Props {
    event: GQL.IReviewRequestedEvent

    className?: string
}

export const ReviewRequestedEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <CampaignTimelineItem icon={PlusCircleIcon} className={className} event={event}>
        <PersonLink user={event.actor} /> requested a review on <Link to={event.thread.url}>{event.thread.title}</Link>
    </CampaignTimelineItem>
)
