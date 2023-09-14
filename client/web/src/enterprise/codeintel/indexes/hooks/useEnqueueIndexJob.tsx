import { type MutationFunctionOptions, type FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import type {
    QueueAutoIndexJobsForRepoVariables,
    Exact,
    Maybe,
    QueueAutoIndexJobsForRepoResult,
} from '../../../../graphql-operations'

import { preciseIndexFieldsFragment } from './types'

const QUEUE_AUTO_INDEX_JOBS = gql`
    mutation QueueAutoIndexJobsForRepo($id: ID!, $rev: String) {
        queueAutoIndexJobsForRepo(repository: $id, rev: $rev) {
            ...PreciseIndexFields
        }
    }

    ${preciseIndexFieldsFragment}
`

type EnqueueIndexJobResults = Promise<
    FetchResult<QueueAutoIndexJobsForRepoResult, Record<string, any>, Record<string, any>>
>
interface UseEnqueueIndexJobResult {
    handleEnqueueIndexJob: (
        options?:
            | MutationFunctionOptions<QueueAutoIndexJobsForRepoResult, Exact<{ id: string; rev: Maybe<string> }>>
            | undefined
    ) => EnqueueIndexJobResults
}

export const useEnqueueIndexJob = (): UseEnqueueIndexJobResult => {
    const [handleEnqueueIndexJob] = useMutation<QueueAutoIndexJobsForRepoResult, QueueAutoIndexJobsForRepoVariables>(
        getDocumentNode(QUEUE_AUTO_INDEX_JOBS)
    )

    return {
        handleEnqueueIndexJob,
    }
}
