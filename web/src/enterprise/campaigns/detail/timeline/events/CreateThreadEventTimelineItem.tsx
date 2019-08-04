import PlusCircleIcon from 'mdi-react/PlusCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { PersonLink } from '../../../../../user/PersonLink'
import { CampaignTimelineItem } from '../CampaignTimelineItem'

interface Props {
    event: GQL.ICreateThreadEvent

    className?: string
}

export const CreateThreadEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => (
    <CampaignTimelineItem icon={PlusCircleIcon} className={className} event={event}>
        <PersonLink user={event.actor} /> added the thread <Link to={event.thread.url}>{event.thread.title}</Link>
    </CampaignTimelineItem>
)
