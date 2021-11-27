import { gql } from '@sourcegraph/shared/src/graphql/graphql'

const CATALOG_COMPONENT_DETAIL_FRAGMENT = gql`
    fragment CatalogComponentDetailFields on CatalogComponent {
        id
        kind
        name
        system
        tags
        sourceLocation {
            url
        }
    }
`

export const CATALOG_COMPONENT_BY_ID = gql`
    query CatalogComponentByID($id: ID!) {
        node(id: $id) {
            __typename
            ... on CatalogComponent {
                ...CatalogComponentDetailFields
            }
        }
    }
    ${CATALOG_COMPONENT_DETAIL_FRAGMENT}
`
