import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const CATALOG_ENTITIES = gql`
    query CatalogEntities($query: String, $first: Int, $after: String) {
        catalog {
            entities(query: $query, first: $first, after: $after) {
                nodes {
                    ...CatalogEntityFields
                }
            }
        }
    }

    fragment CatalogEntityFields on CatalogEntity {
        __typename
        id
        type
        name
        url
        ... on CatalogComponent {
            kind
        }
    }
`
