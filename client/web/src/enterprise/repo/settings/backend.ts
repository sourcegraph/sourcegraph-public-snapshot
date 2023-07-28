import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'

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

export const useExternalServicesConnection = (
    vars: ExternalServicesVariables
): UseShowMorePaginationResult<ExternalServicesResult, ListExternalServiceFields> =>
    useShowMorePagination<ExternalServicesResult, ExternalServicesVariables, ListExternalServiceFields>({
        query: EXTERNAL_SERVICES,
        variables: { after: vars.after, first: vars.first ?? 10 },
        getConnection: result => {
            const { externalServices } = dataOrThrowErrors(result)
            return externalServices
        },
        options: {
            fetchPolicy: 'cache-and-network',
            pollInterval: 15000,
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)

            if (!node) {
                throw new Error('Repository not found')
            }
        },
    })
