import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ActorLink } from '../../actor/ActorLink'

export interface ParticipantListItemContext {
    showRepository?: boolean
}

interface Props extends ParticipantListItemContext {
    participant: GQL.IParticipantEdge

    className?: string
}

/**
 * An item in a list of participants.
 */
export const ParticipantListItem: React.FunctionComponent<Props> = ({ participant, className = '' }) => (
    <li className={`list-group-item d-flex align-items-center ${className}`}>
        <ActorLink actor={participant.actor} />
    </li>
)
