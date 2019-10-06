import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CircleIcon from 'mdi-react/CircleIcon'
import CommaCircleIcon from 'mdi-react/CommaCircleIcon'
import PencilCircleIcon from 'mdi-react/PencilCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.IReviewEvent

    className?: string
}

export const ReviewEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => {
    const { icon, className: iconClassName } = stateIcon(event.state)
    return (
        <TimelineItem icon={icon} className={className} iconClassName={iconClassName} event={event}>
            <ActorLink actor={event.actor} /> {stateDescription(event.state)}{' '}
            <Link to={event.thread.url}>{event.thread.title}</Link>
        </TimelineItem>
    )
}

function stateIcon(state: GQL.ReviewState): { icon: React.ComponentType<{ className?: string }>; className: string } {
    switch (state) {
        case GQL.ReviewState.COMMENTED:
            return { icon: CommaCircleIcon, className: 'text-muted' }
        case GQL.ReviewState.CHANGES_REQUESTED:
            return { icon: PencilCircleIcon, className: 'text-warning' }
        case GQL.ReviewState.APPROVED:
            return { icon: CheckCircleIcon, className: 'text-success' }
        default:
            return { icon: CircleIcon, className: '' }
    }
}

function stateDescription(state: GQL.ReviewState): string {
    switch (state) {
        case GQL.ReviewState.COMMENTED:
            return 'submitted review comments on'
        case GQL.ReviewState.CHANGES_REQUESTED:
            return 'requested changes on'
        case GQL.ReviewState.APPROVED:
            return 'approved'
        default:
            return 'reviewed'
    }
}
