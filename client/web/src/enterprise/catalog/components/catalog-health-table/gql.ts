import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_ENTITY_STATE_FRAGMENT } from '../../pages/overview/components/catalog-explorer/gql'
import { CATALOG_ENTITY_OWNER_FRAGMENT } from '../entity-owner/gql'

const CATALOG_HEALTH_FRAGMENT = gql`
    fragment CatalogEntityHealthFields on CatalogEntity {
        __typename
        id
        type
        name
        url
        ... on CatalogComponent {
            kind
        }
        status {
            id
            contexts {
                id
                name
                state
                title
                description
                targetURL
            }
        }
        ...CatalogEntityOwnerFields
        ...CatalogEntityStateFields
    }
    ${CATALOG_ENTITY_OWNER_FRAGMENT}
    ${CATALOG_ENTITY_STATE_FRAGMENT}
`

export const CATALOG_HEALTH = gql`
    query CatalogHealth($query: String, $first: Int, $after: String) {
        catalog {
            entities(query: $query, first: $first, after: $after) {
                nodes {
                    ...CatalogEntityHealthFields
                }
            }
        }
    }
    ${CATALOG_HEALTH_FRAGMENT}
`
