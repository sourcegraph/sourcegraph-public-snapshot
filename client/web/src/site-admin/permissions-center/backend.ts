import { gql } from '@sourcegraph/http-client'

export const PERMISSIONS_SYNC_JOBS_QUERY = gql`
    fragment CodeHostState on CodeHostState {
        providerID
        providerType
        status
    }

    fragment PermissionsSyncJob on PermissionsSyncJob {
        id
        state
        subject {
            ... on Repository {
                __typename
                name
                externalRepository {
                    serviceType
                }
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
            reason
        }
        queuedAt
        startedAt
        finishedAt
        processAfter
        permissionsAdded
        permissionsRemoved
        permissionsFound
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
            ...CodeHostState
        }
    }

    query PermissionsSyncJobs(
        $first: Int
        $last: Int
        $after: String
        $before: String
        $reasonGroup: PermissionsSyncJobReasonGroup
        $state: PermissionsSyncJobState
        $searchType: PermissionsSyncJobsSearchType
        $query: String
    ) {
        permissionsSyncJobs(
            first: $first
            last: $last
            after: $after
            before: $before
            reasonGroup: $reasonGroup
            state: $state
            searchType: $searchType
            query: $query
        ) {
            totalCount
            pageInfo {
                hasNextPage
                hasPreviousPage
                startCursor
                endCursor
            }
            nodes {
                ...PermissionsSyncJob
            }
        }
    }
`
