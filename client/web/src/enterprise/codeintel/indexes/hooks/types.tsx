import { gql } from '@sourcegraph/http-client'

export const lsifIndexFieldsFragment = gql`
    fragment LsifIndexFields on LSIFIndex {
        __typename
        id
        inputCommit
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
