import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'

interface Props {
    actor: Pick<GQL.Actor, '__typename' | 'username' | 'displayName' | 'url'> | null
}

/**
 * A link to an actor.
 */
export const ActorLink: React.FunctionComponent<Props> = ({ actor }) =>
    actor ? <Link to={actor.url}>{actor.username}</Link> : <span className="font-style-italic">unknown actor</span>
