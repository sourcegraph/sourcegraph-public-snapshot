import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { RouterLinkOrAnchor } from '../components/RouterLinkOrAnchor'

interface Props {
    actor: Pick<GQL.Actor, '__typename' | 'username' | 'displayName' | 'url'> | null
}

/**
 * A link to an actor.
 */
export const ActorLink: React.FunctionComponent<Props> = ({ actor }) =>
    actor ? (
        <RouterLinkOrAnchor to={actor.url}>{actor.username}</RouterLinkOrAnchor>
    ) : (
        <span className="font-style-italic">unknown actor</span>
    )
