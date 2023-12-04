import { gql } from '@sourcegraph/http-client'

export const PERMISSIONS_SYNC_JOBS_QUERY = gql`
    fragment CodeHostState on CodeHostState {
        providerID
        providerType
        status
        message
    }

    fragment PermissionsSyncJob on PermissionsSyncJob {
        id
        state
        subject {
            ... on Repository {
                __typename
                id
                name
                url
                externalRepository {
                    serviceType
                    serviceID
                }
            }
            ... on User {
                __typename
                id
                username
                displayName
                email
                avatarURL
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
        placeInQueue
        codeHostStates {
            ...CodeHostState
        }
        partialSuccess
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
        $userID: ID
        $repoID: ID
        $partial: Boolean
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
            userID: $userID
            repoID: $repoID
            partial: $partial
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

export const TRIGGER_USER_SYNC = gql`
    mutation ScheduleUserPermissionsSync($user: ID!) {
        scheduleUserPermissionsSync(user: $user) {
            alwaysNil
        }
    }
`

export const TRIGGER_REPO_SYNC = gql`
    mutation ScheduleRepoPermissionsSync($repo: ID!) {
        scheduleRepositoryPermissionsSync(repository: $repo) {
            alwaysNil
        }
    }
`

export const CANCEL_PERMISSIONS_SYNC_JOB = gql`
    mutation CancelPermissionsSyncJob($job: ID!) {
        cancelPermissionsSyncJob(job: $job)
    }
`

export const PERMISSIONS_SYNC_JOBS_STATS = gql`
    query PermissionsSyncJobsStats {
        permissionsSyncingStats {
            queueSize
            usersWithLatestJobFailing
            reposWithLatestJobFailing
            usersWithNoPermissions
            reposWithNoPermissions
            usersWithStalePermissions
            reposWithStalePermissions
        }
        site {
            users(deletedAt: { empty: true }) {
                totalCount
            }
        }
        repositoryStats {
            total
        }
    }
`
