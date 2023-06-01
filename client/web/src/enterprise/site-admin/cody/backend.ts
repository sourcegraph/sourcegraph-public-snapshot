import { MutationTuple } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../backend/graphql'
import { FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import {
    useShowMorePagination,
    UseShowMorePaginationResult,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    RepoEmbeddingJobFields,
    RepoEmbeddingJobsListResult,
    RepoEmbeddingJobsListVariables,
    RepoEmbeddingJobConnectionFields,
    ScheduleContextDetectionEmbeddingJobResult,
    ScheduleContextDetectionEmbeddingJobVariables,
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
        repo {
            name
            url
        }
        revision {
            oid
            abbreviatedOID
        }
    }
`

const REPO_EMBEDDING_JOB_CONNECTION_FIELDS_FRAGMENT = gql`
    ${REPO_EMBEDDING_JOB_FRAGMENT}
    fragment RepoEmbeddingJobConnectionFields on RepoEmbeddingJobsConnection {
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
        nodes {
            ...RepoEmbeddingJobFields
        }
    }
`

export const REPO_EMBEDDING_JOBS_LIST_QUERY = gql`
    ${REPO_EMBEDDING_JOB_CONNECTION_FIELDS_FRAGMENT}
    query RepoEmbeddingJobsList($first: Int, $after: String) {
        repoEmbeddingJobs(first: $first, after: $after) {
            ...RepoEmbeddingJobConnectionFields
        }
    }
`

export function repoEmbeddingJobs(
    variables: FilteredConnectionQueryArguments
): Observable<RepoEmbeddingJobConnectionFields> {
    return requestGraphQL<RepoEmbeddingJobsListResult, RepoEmbeddingJobsListVariables>(REPO_EMBEDDING_JOBS_LIST_QUERY, {
        first: variables.first ?? null,
        after: variables.after ?? null,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.repoEmbeddingJobs)
    )
}

export const useRepoEmbeddingJobsConnection = (): UseShowMorePaginationResult<
    RepoEmbeddingJobsListResult,
    RepoEmbeddingJobFields
> =>
    useShowMorePagination<RepoEmbeddingJobsListResult, RepoEmbeddingJobsListVariables, RepoEmbeddingJobFields>({
        query: REPO_EMBEDDING_JOBS_LIST_QUERY,
        variables: {
            after: null,
            first: 10,
        },
        getConnection: result => {
            const { repoEmbeddingJobs } = dataOrThrowErrors(result)
            return repoEmbeddingJobs
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

export const SCHEDULE_CONTEXT_DETECTION_EMBEDDING_JOB = gql`
    mutation ScheduleContextDetectionEmbeddingJob {
        scheduleContextDetectionForEmbedding {
            alwaysNil
        }
    }
`

export function useScheduleContextDetectionEmbeddingJob(): MutationTuple<
    ScheduleContextDetectionEmbeddingJobResult,
    ScheduleContextDetectionEmbeddingJobVariables
> {
    return useMutation<ScheduleContextDetectionEmbeddingJobResult, ScheduleContextDetectionEmbeddingJobVariables>(
        SCHEDULE_CONTEXT_DETECTION_EMBEDDING_JOB
    )
}
