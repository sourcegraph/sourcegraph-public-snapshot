import { gql } from '@sourcegraph/http-client'

// TODO(sashaostrikov): move all properties to PermissionsSyncJob and make a storybook more elaborate
export const PERMISSIONS_SYNC_JOBS_QUERY = gql`
    fragment PermissionsSyncJob on PermissionSyncJob {
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

    query PermissionSyncJobs($first: Int, $last: Int, $after: String, $before: String) {
        permissionSyncJobs(first: $first, last: $last, after: $after, before: $before) {
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
