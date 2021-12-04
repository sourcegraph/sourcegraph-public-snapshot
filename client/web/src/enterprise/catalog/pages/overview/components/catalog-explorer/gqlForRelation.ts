import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_ENTITY_FOR_EXPLORER_FRAGMENT } from './gql'

export const CATALOG_ENTITY_RELATIONS_FOR_EXPLORER = gql`
    query CatalogEntityRelationsForExplorer($entity: ID!, $query: String, $first: Int, $after: String) {
        node(id: $entity) {
            __typename
            ... on CatalogEntity {
                relatedEntities(query: $query, first: $first, after: $after) {
                    edges {
                        ...CatalogEntityRelationFields
                    }
                }
            }
        }
    }

    fragment CatalogEntityRelationFields on CatalogEntityRelatedEntityEdge {
        type
        node {
            ...CatalogEntityForExplorerFields
        }
    }

    ${CATALOG_ENTITY_FOR_EXPLORER_FRAGMENT}
`
