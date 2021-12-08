import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_ENTITY_OWNER_FRAGMENT } from '../../../../components/entity-owner/gql'

export const CATALOG_ENTITY_LAST_COMMIT_FRAGMENT = gql`
    fragment CatalogEntityLastCommitFields on CatalogEntity {
        ... on CatalogComponent {
            commits(first: 1) {
                nodes {
                    author {
                        date
                    }
                }
            }
        }
    }
`

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
        url
        ... on CatalogComponent {
            kind
            lifecycle
        }
        ...CatalogEntityStateFields
        ...CatalogEntityOwnerFields
        ...CatalogEntityLastCommitFields
    }
    ${CATALOG_ENTITY_STATE_FRAGMENT}
    ${CATALOG_ENTITY_OWNER_FRAGMENT}
    ${CATALOG_ENTITY_LAST_COMMIT_FRAGMENT}
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
