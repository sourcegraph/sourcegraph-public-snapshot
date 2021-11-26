import { gql } from '@sourcegraph/http-client'

export const CATALOG_GRAPH_FRAGMENT = gql`
    fragment CatalogGraphFields on CatalogGraph {
        nodes {
            __typename
            id
            name
            kind
            description
            url
        }
        edges {
            type
            outNode {
                id
            }
            inNode {
                id
            }
        }
    }
`
