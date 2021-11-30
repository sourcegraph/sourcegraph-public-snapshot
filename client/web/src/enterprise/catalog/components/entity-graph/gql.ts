import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const CATALOG_GRAPH_FRAGMENT = gql`
    fragment CatalogGraphFields on CatalogGraph {
        nodes {
            __typename
            id
            type
            name
            description
            url
            ... on CatalogComponent {
                kind
            }
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
