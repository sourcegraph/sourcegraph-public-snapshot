import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_ENTITY_OWNER_FRAGMENT } from '../../../../components/entity-owner/gql'

export const CATALOG_ENTITY_STATE_FRAGMENT = gql`
    fragment CatalogEntityStateFields on CatalogEntity {
        status {
            id
            state
        }
    }
`

export const CATALOG_ENTITY_FOR_EXPLORER_FRAGMENT = gql`
    fragment CatalogEntityForExplorerFields on CatalogEntity {
        __typename
        id
        type
        name
        description
        lifecycle
        url
        ... on CatalogComponent {
            kind
        }
        ...CatalogEntityStateFields
        ...CatalogEntityOwnerFields
    }
    ${CATALOG_ENTITY_STATE_FRAGMENT}
    ${CATALOG_ENTITY_OWNER_FRAGMENT}
`

export const CATALOG_ENTITIES_FOR_EXPLORER = gql`
    query CatalogEntitiesForExplorer($query: String, $first: Int, $after: String) {
        catalog {
            entities(query: $query, first: $first, after: $after) {
                nodes {
                    ...CatalogEntityForExplorerFields
                }
            }
        }
    }
    ${CATALOG_ENTITY_FOR_EXPLORER_FRAGMENT}
`
