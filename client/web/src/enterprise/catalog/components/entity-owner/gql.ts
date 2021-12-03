import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'

export const CATALOG_ENTITY_OWNER_FRAGMENT = gql`
    fragment CatalogEntityOwnerFields on CatalogEntity {
        owner {
            __typename
            ... on Person {
                ...PersonLinkFields
            }
            ... on Group {
                id
                name
                title
                url
            }
        }
    }
    ${personLinkFieldsFragment}
`
