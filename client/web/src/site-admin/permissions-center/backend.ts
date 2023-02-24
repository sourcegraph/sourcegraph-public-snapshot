import { gql } from '@sourcegraph/http-client'

export const PERMISSIONS_SYNC_JOBS_QUERY = gql`
    fragment PermissionsSyncJob on PermissionsSyncJob {
        id
        state
        subject {
            ... on Repository {
                __typename
                name
            }
            ... on User {
                __typename
                username
            }
        }
        triggeredByUser {
            username
        }
        reason {
            group
            message
        }
        queuedAt
        startedAt
        finishedAt
        processAfter
        permissionsAdded
        permissionsRemoved
        permissionsFound
    }

    query PermissionsSyncJobs($first: Int, $last: Int, $after: String, $before: String) {
        permissionsSyncJobs(first: $first, last: $last, after: $after, before: $before) {
            totalCount
            pageInfo {
                hasNextPage
                hasPreviousPage
                startCursor
                endCursor
            }
            nodes {
                ...PermissionsSyncJob
                failureMessage
                cancellationReason
                ranForMs
                numResets
                numFailures
                lastHeartbeatAt
                workerHostname
                cancel
                priority
                noPerms
                invalidateCaches
                codeHostStates {
                    providerID
                }
            }
        }
    }
`
