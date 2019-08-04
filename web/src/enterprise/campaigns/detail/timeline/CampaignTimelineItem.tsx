import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../../components/time/Timestamp'

interface Props {
    icon: React.ComponentType<{ className?: string }>
    event: Pick<GQL.EventCommon, 'id' | 'createdAt'>

    tag?: 'li' | 'div'
    className?: string
}

export const CampaignTimelineItem: React.FunctionComponent<Props> = ({
    icon: Icon,
    event,
    tag: Tag = 'li',
    className = '',
    children,
}) => (
    <Tag className={`d-flex align-items-start ml-4 pl-1 ${className}`} id={`event-${event.id}`}>
        <Icon className="icon-inline mr-3 text-muted" />
        <div>
            {children}{' '}
            <span className="text-muted">
                &mdash; <Timestamp date={event.createdAt} />
            </span>
        </div>
    </Tag>
)
