import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Timestamp } from '../../components/time/Timestamp'

interface Props {
    icon: React.ComponentType<{ className?: string }>
    event: Pick<GQL.EventCommon, 'id' | 'createdAt'>

    tag?: 'li' | 'div'
    className?: string
    iconClassName?: string
}

/**
 * A timeline item (representing an event) in a timeline.
 */
export const TimelineItem: React.FunctionComponent<Props> = ({
    icon: Icon,
    event,
    tag: Tag = 'li',
    className = '',
    iconClassName = 'text-muted',
    children,
}) => (
    <Tag className={`d-flex align-items-start ml-4 pl-1 ${className}`} id={`event-${event.id}`}>
        <Icon className={`icon-inline mr-3 ${iconClassName}`} />
        <div>
            {children}{' '}
            <span className="text-muted">
                &mdash; <Timestamp date={event.createdAt} />
            </span>
        </div>
    </Tag>
)
