import * as GQL from '../../../shared/src/graphql/schema'

/**
 * Returns the actor username or similar (e.g., the organization name for an organization actor).
 */
export const actorName = (actor: GQL.Actor): string => {
    switch (actor.__typename) {
        case 'User':
            return actor.username
        case 'Org':
            return actor.name
        case 'ExternalActor':
            return actor.username
    }
}
