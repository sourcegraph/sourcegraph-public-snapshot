import { gql } from '@sourcegraph/http-client'

export const codeIntelIndexerFieldsFragment = gql`
    fragment CodeIntelIndexerFields on CodeIntelIndexer {
        key
        name
        url
    }
`

export const preciseIndexFieldsFragment = gql`
    fragment PreciseIndexFields on PreciseIndex {
        __typename
        id
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
        inputCommit
        tags
        inputRoot
        inputIndexer
        indexer {
            ...CodeIntelIndexerFields
        }
        state
        queuedAt
        uploadedAt
        indexingStartedAt
        indexingFinishedAt
        processingStartedAt
        processingFinishedAt
        steps {
            ...LsifIndexStepsFields
        }
        failure
        placeInQueue
        shouldReindex
        isLatestForRepo

        auditLogs {
            ...PreciseIndexAuditLogFields
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

    fragment PreciseIndexAuditLogFields on LSIFUploadAuditLog {
        logTimestamp
        reason
        changedColumns {
            column
            old
            new
        }
        operation
    }

    ${codeIntelIndexerFieldsFragment}
`

export const preciseIndexConnectionFieldsFragment = gql`
    fragment PreciseIndexConnectionFields on PreciseIndexConnection {
        nodes {
            ...PreciseIndexFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }

    ${preciseIndexFieldsFragment}
`
