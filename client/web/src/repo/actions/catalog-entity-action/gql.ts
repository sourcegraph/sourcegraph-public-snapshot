import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_ENTITY_DETAIL_FRAGMENT } from '../../../enterprise/catalog/pages/entity-detail/global/gql'
import { CATALOG_ENTITY_STATE_FRAGMENT } from '../../../enterprise/catalog/pages/overview/components/catalog-explorer/gql'

// TODO(sqs): only works for blobs not trees right now

export const TREE_ENTRY_CATALOG_ENTITY = gql`
    query TreeEntryCatalogEntity($repository: ID!, $rev: String!, $path: String!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                commit(rev: $rev) {
                    blob(path: $path) {
                        ...TreeEntryCatalogEntityFields
                    }
                }
            }
        }
    }

    fragment TreeEntryCatalogEntityFields on TreeEntry {
        catalogEntities {
            __typename
            id
            type
            name
            description
            url
            ... on CatalogComponent {
                kind
            }
            ...CatalogEntityStateFields
            ...CatalogEntityDetailFields
        }
    }

    ${CATALOG_ENTITY_STATE_FRAGMENT}
    ${CATALOG_ENTITY_DETAIL_FRAGMENT}
`
