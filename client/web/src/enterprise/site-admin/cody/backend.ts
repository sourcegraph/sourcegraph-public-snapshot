import type { MutationTuple } from '@apollo/client'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    CancelRepoEmbeddingJobResult,
    CancelRepoEmbeddingJobVariables,
    RepoEmbeddingJobFields,
    RepoEmbeddingJobsListResult,
    RepoEmbeddingJobsListVariables,
    ScheduleRepoEmbeddingJobsResult,
    ScheduleRepoEmbeddingJobsVariables,
} from '../../../graphql-operations'

const REPO_EMBEDDING_JOB_FRAGMENT = gql`
    fragment RepoEmbeddingJobFields on RepoEmbeddingJob {
        id
        state
        failureMessage
        finishedAt
        queuedAt
        startedAt
        cancel
        repo {
            name
            url
        }
        revision {
            oid
            abbreviatedOID
        }
        stats {
            filesScheduled
            filesEmbedded
            filesSkipped
        }
    }
`

export const REPO_EMBEDDING_JOBS_LIST_QUERY = gql`
    ${REPO_EMBEDDING_JOB_FRAGMENT}

    query RepoEmbeddingJobsList($first: Int, $after: String, $query: String, $state: String) {
        repoEmbeddingJobs(first: $first, after: $after, query: $query, state: $state) {
            nodes {
                ...RepoEmbeddingJobFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }
`

export const useRepoEmbeddingJobsConnection = (
    query: string,
    state: string | null
): UseShowMorePaginationResult<RepoEmbeddingJobsListResult, RepoEmbeddingJobFields> =>
    useShowMorePagination<RepoEmbeddingJobsListResult, RepoEmbeddingJobsListVariables, RepoEmbeddingJobFields>({
        query: REPO_EMBEDDING_JOBS_LIST_QUERY,
        variables: { after: null, first: 10, query, state },
        getConnection: result => {
            const { repoEmbeddingJobs } = dataOrThrowErrors(result)
            return repoEmbeddingJobs
        },
        options: {
            pollInterval: 5000,
        },
    })

export const SCHEDULE_REPO_EMBEDDING_JOBS = gql`
    mutation ScheduleRepoEmbeddingJobs($repoNames: [String!]!) {
        scheduleRepositoriesForEmbedding(repoNames: $repoNames) {
            alwaysNil
        }
    }
`

export function useScheduleRepoEmbeddingJobs(): MutationTuple<
    ScheduleRepoEmbeddingJobsResult,
    ScheduleRepoEmbeddingJobsVariables
> {
    return useMutation<ScheduleRepoEmbeddingJobsResult, ScheduleRepoEmbeddingJobsVariables>(
        SCHEDULE_REPO_EMBEDDING_JOBS
    )
}

export const CANCEL_REPO_EMBEDDING_JOB = gql`
    mutation CancelRepoEmbeddingJob($id: ID!) {
        cancelRepoEmbeddingJob(job: $id) {
            alwaysNil
        }
    }
`

export function useCancelRepoEmbeddingJob(): MutationTuple<
    CancelRepoEmbeddingJobResult,
    CancelRepoEmbeddingJobVariables
> {
    return useMutation<CancelRepoEmbeddingJobResult, CancelRepoEmbeddingJobVariables>(CANCEL_REPO_EMBEDDING_JOB)
}
