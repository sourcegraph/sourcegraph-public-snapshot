import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { COMPONENT_OWNER_FRAGMENT } from '../../../enterprise/catalog/components/entity-owner/gql'
import {
    COMPONENT_AUTHORS_FRAGMENT,
    COMPONENT_CODE_OWNERS_FRAGMENT,
    COMPONENT_STATUS_FRAGMENT,
    COMPONENT_USAGE_PEOPLE_FRAGMENT,
} from '../../../enterprise/catalog/pages/component/gql'

export const COMPONENTS_FOR_TREE_ENTRY_HEADER_ACTION = gql`
    query ComponentsForTreeEntryHeaderAction($repository: ID!, $rev: String!, $path: String!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                id
                commit(rev: $rev) {
                    id
                    treeEntry(path: $path) {
                        id
                        ...ComponentsForTreeEntryHeaderActionFields
                    }
                }
            }
        }
    }

    fragment ComponentsForTreeEntryHeaderActionFields on TreeEntry {
        components {
            __typename
            id
            name
            kind
            description
            lifecycle
            url
            ...ComponentOwnerFields
            ...ComponentStatusFields
            ...ComponentCodeOwnersFields
            ...ComponentAuthorsFields
            ...ComponentUsagePeopleFields
        }
    }

    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_STATUS_FRAGMENT}
    ${COMPONENT_CODE_OWNERS_FRAGMENT}
    ${COMPONENT_AUTHORS_FRAGMENT}
    ${COMPONENT_USAGE_PEOPLE_FRAGMENT}
`
