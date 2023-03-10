import { gql } from '@sourcegraph/http-client'

export const RepoPermissionsInfoQuery = gql`
    query RepoPermissionsInfo($repoID: ID!, $first: Int, $last: Int, $after: String, $before: String, $query: String) {
        node(id: $repoID) {
            __typename
            ... on Repository {
                ...RepoPermissionsInfoRepoNode
            }
        }
    }

    fragment RepoPermissionsInfoRepoNode on Repository {
        permissionsInfo {
            syncedAt
            updatedAt
            unrestricted
            users(first: $first, last: $last, after: $after, before: $before, query: $query) {
                nodes {
                    ...PermissionsInfoUserFields
                }
                totalCount
                pageInfo {
                    hasNextPage
                    hasPreviousPage
                    startCursor
                    endCursor
                }
            }
        }
    }

    fragment PermissionsInfoUserFields on PermissionsInfoUserNode {
        id
        reason
        updatedAt
        user {
            id
            username
            displayName
            email
            avatarURL
        }
    }
`
