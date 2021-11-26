import { useQuery, gql } from '@sourcegraph/http-client'

import { COMPONENT_OWNER_FRAGMENT } from '../../../../components/entity-owner/gql'

export const COMPONENT_LAST_COMMIT_FRAGMENT = gql`
    fragment ComponentLastCommitFields on Component {
        commits(first: 1) {
            nodes {
                author {
                    date
                }
            }
        }
    }
`

export const COMPONENT_STATE_FRAGMENT = gql`
    fragment ComponentStateFields on Component {
        status {
            id
            state
        }
    }
`

export const COMPONENT_FOR_EXPLORER_FRAGMENT = gql`
    fragment ComponentForExplorerFields on Component {
        __typename
        id
        name
        kind
        description
        lifecycle
        url
        catalogURL
        ...ComponentStateFields
        ...ComponentOwnerFields
        ...ComponentLastCommitFields
    }
    ${COMPONENT_STATE_FRAGMENT}
    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_LAST_COMMIT_FRAGMENT}
`

export const COMPONENTS_FOR_EXPLORER = gql`
    query ComponentsForExplorer($query: String, $first: Int, $after: String) {
        components(query: $query, first: $first, after: $after) {
            nodes {
                ...ComponentForExplorerFields
            }
            totalCount
            pageInfo {
                hasNextPage
            }
            tags {
                name
            }
        }
    }
    ${COMPONENT_FOR_EXPLORER_FRAGMENT}
`
