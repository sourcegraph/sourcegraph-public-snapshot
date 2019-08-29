import { Diagnostic } from '@sourcegraph/extension-api-types'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ActorLink } from '../../../actor/ActorLink'
import { DiagnosticSeverityIcon } from '../../../diagnostics/components/DiagnosticSeverityIcon'
import { TimelineItem } from '../TimelineItem'

interface Props {
    event: GQL.IAddDiagnosticToThreadEvent | GQL.IRemoveDiagnosticFromThreadEvent

    className?: string
}

export const ThreadDiagnosticEdgeEventTimelineItem: React.FunctionComponent<Props> = ({ event, className = '' }) => {
    const diagnostic: Diagnostic = event.diagnostic.data
    return (
        <TimelineItem icon={AlertCircleOutlineIcon} className={`${className}`} event={event}>
            <ActorLink actor={event.actor} /> {event.__typename === 'AddDiagnosticToThreadEvent' ? 'added' : 'removed'}{' '}
            the diagnostic <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline" />{' '}
            <Link to={`${event.thread.url}/diagnostics`}>{diagnostic.message}</Link>{' '}
            {event.__typename === 'AddDiagnosticToThreadEvent' ? 'to' : 'from'}{' '}
            <Link to={event.thread.url}>{event.thread.title}</Link>
        </TimelineItem>
    )
}
