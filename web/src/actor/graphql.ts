import { gql } from '../../../shared/src/graphql/graphql'

export const ActorQuery = gql`
__typename
... on User {
    username
    displayName
    url
}
... on Org {
    name
    displayName
    url
}
... on ExternalActor {
    username
    displayName
    url
}
`

export const ActorFragment = gql``
