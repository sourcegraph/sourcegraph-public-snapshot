import CloseCircleIcon from 'mdi-react/CloseCircleIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.IRequestReviewEvent

    className?: string
}

export const CloseThreadEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) =>
    event.thread.state !== GQL.ThreadState.MERGED ? (
        <TimelineItem
            icon={CloseCircleIcon}
            className={className}
            iconClassName={event.thread.kind === GQL.ThreadKind.CHANGESET ? 'text-danger' : 'text-success'}
            event={event}
        >
            <ActorLink actor={event.actor} /> closed <Link to={event.thread.url}>{event.thread.title}</Link>
        </TimelineItem>
    ) : null
