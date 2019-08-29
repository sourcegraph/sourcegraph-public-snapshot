import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { RouterLinkOrAnchor } from '../components/RouterLinkOrAnchor'
import { actorName } from '.'

interface Props {
    actor: GQL.Actor | null

    className?: string
}

/**
 * A link to an actor.
 */
export const ActorLink: React.FunctionComponent<Props> = ({ actor, className = '' }) =>
    actor ? (
        <RouterLinkOrAnchor to={actor.url} className={className}>
            {actorName(actor)}
        </RouterLinkOrAnchor>
    ) : (
        <span className={`font-style-italic ${className}`}>unknown actor</span>
    )
