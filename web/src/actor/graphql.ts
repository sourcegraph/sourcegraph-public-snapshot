import { queryAndFragmentForUnion } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'

export const { query: actorQuery, fragment: actorFragment } = queryAndFragmentForUnion<
    GQL.Actor['__typename'],
    keyof GQL.Actor
>(['User', 'Org', 'ExternalActor'], ['__typename', 'username', 'displayName', 'url'])
