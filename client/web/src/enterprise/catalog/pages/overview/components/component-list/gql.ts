import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const CATALOG_COMPONENTS = gql`
    query CatalogComponents($query: String, $first: Int, $after: String) {
        catalog {
            components(query: $query, first: $first, after: $after) {
                nodes {
                    ...CatalogComponentFields
                }
            }
        }
    }

    fragment CatalogComponentFields on CatalogComponent {
        id
        kind
        name
        system
        tags
        url
    }
`
