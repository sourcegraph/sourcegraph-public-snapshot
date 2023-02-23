import { ApolloError, gql, useQuery } from '@apollo/client'

import { InferAutoIndexJobsForRepoVariables, InferAutoIndexJobsForRepoResult } from '../../../../graphql-operations'

interface InferJobsScriptResult {
    data?: InferAutoIndexJobsForRepoResult
    loading: boolean
    error: ApolloError | undefined
}

export const INFER_JOBS_SCRIPT = gql`
    query InferAutoIndexJobsForRepo($repository: ID!, $rev: String, $script: String) {
        inferAutoIndexJobsForRepo(repository: $repository, rev: $rev, script: $script) {
            ...AutoIndexJobDescriptionFields
        }
    }

    fragment AutoIndexJobDescriptionFields on AutoIndexJobDescription {
        root
        indexer {
            name
            url
        }
        steps {
            ...LsifIndexStepsFields
        }
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

export const useInferJobs = ({
    variables,
}: {
    variables: InferAutoIndexJobsForRepoVariables
}): InferJobsScriptResult => {
    const { data, loading, error } = useQuery<InferAutoIndexJobsForRepoResult>(INFER_JOBS_SCRIPT, {
        variables,
        nextFetchPolicy: 'cache-first',
    })

    return {
        data,
        loading,
        error,
    }
}
