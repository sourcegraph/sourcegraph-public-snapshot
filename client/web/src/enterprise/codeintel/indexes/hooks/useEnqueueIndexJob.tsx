import { MutationFunctionOptions, FetchResult, useMutation } from '@apollo/client'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import {
    QueueAutoIndexJobsForRepoVariables,
    Exact,
    Maybe,
    QueueAutoIndexJobsForRepoResult,
} from '../../../../graphql-operations'

export const lsifIndexFieldsFragment = gql`
    fragment LsifIndexFields on LSIFIndex {
        __typename
        id
        inputCommit
        tags
        inputRoot
        inputIndexer
        indexer {
            name
            url
        }
        projectRoot {
            url
            path
            repository {
                url
                name
            }
            commit {
                url
                oid
                abbreviatedOID
            }
        }
        steps {
            ...LsifIndexStepsFields
        }
        state
        failure
        queuedAt
        startedAt
        finishedAt
        placeInQueue
        associatedUpload {
            id
            state
            uploadedAt
            startedAt
            finishedAt
            placeInQueue
        }
        shouldReindex
    }

    fragment LsifIndexStepsFields on IndexSteps {
        setup {
            ...ExecutionLogEntryFields
        }
        preIndex {
            root
            image
            commands
            logEntry {
                ...ExecutionLogEntryFields
            }
        }
        index {
            indexerArgs
            outfile
            logEntry {
                ...ExecutionLogEntryFields
            }
        }
        upload {
            ...ExecutionLogEntryFields
        }
        teardown {
            ...ExecutionLogEntryFields
        }
    }

    fragment ExecutionLogEntryFields on ExecutionLogEntry {
        key
        command
        startTime
        exitCode
        out
        durationMilliseconds
    }
`

const QUEUE_AUTO_INDEX_JOBS = gql`
    mutation QueueAutoIndexJobsForRepo($id: ID!, $rev: String) {
        queueAutoIndexJobsForRepo(repository: $id, rev: $rev) {
            ...LsifIndexFields
        }
    }

    ${lsifIndexFieldsFragment}
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
