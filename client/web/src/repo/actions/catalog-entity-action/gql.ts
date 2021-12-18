import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { COMPONENT_DETAIL_FRAGMENT } from '../../../enterprise/catalog/pages/component/gql'
import { COMPONENT_STATE_FRAGMENT } from '../../../enterprise/catalog/pages/overview/components/catalog-explorer/gql'

// TODO(sqs): only works for blobs not trees right now

export const TREE_ENTRY_CATALOG_ENTITY = gql`
    query TreeEntryComponent($repository: ID!, $rev: String!, $path: String!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                commit(rev: $rev) {
                    blob(path: $path) {
                        ...TreeEntryComponentsFields
                    }
                }
            }
        }
    }

    fragment TreeEntryComponentsFields on TreeEntry {
        components {
            __typename
            id
            name
            kind
            description
            lifecycle
            url
            ...ComponentStateFields
            ...ComponentStateDetailFields
        }
    }

    ${COMPONENT_STATE_FRAGMENT}
    ${COMPONENT_DETAIL_FRAGMENT}
`
