import { gql } from '@sourcegraph/http-client'

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
            name
            url
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
`

export const globalCodeIntelStatusQuery = gql`
    query GlobalCodeIntelStatus {
        codeIntelSummary {
            numRepositoriesWithCodeIntelligence
            repositoriesWithErrors {
                repository {
                    name
                    url
                    externalRepository {
                        serviceID
                        serviceType
                    }
                }
                count
            }
            repositoriesWithConfiguration {
                repository {
                    name
                    url
                    externalRepository {
                        serviceID
                        serviceType
                    }
                }
                indexers {
                    indexer {
                        ...CodeIntelIndexerFields
                    }
                    count
                }
            }
        }
    }

    fragment CodeIntelIndexerFields on CodeIntelIndexer {
        name
        url
    }
`

export const repoCodeIntelStatusQuery = gql`
    query RepoCodeIntelStatus($repository: String!) {
        repository(name: $repository) {
            codeIntelSummary {
                lastIndexScan
                lastUploadRetentionScan
                recentActivity {
                    ...PreciseIndexFields
                }
                availableIndexers {
                    ...InferredAvailableIndexersFields
                }
            }
            commit(rev: "HEAD") {
                path(path: "/") {
                    ... on GitTree {
                        codeIntelInfo {
                            ...GitTreeCodeIntelInfoFields
                        }
                    }
                }
            }
        }
    }

    fragment InferredAvailableIndexersFields on InferredAvailableIndexers {
        roots
        index
        url
    }

    fragment GitTreeCodeIntelInfoFields on GitTreeCodeIntelInfo {
        preciseSupport {
            support {
                ...PreciseSupportFields
            }
            confidence
        }
        searchBasedSupport {
            support {
                ...SearchBasedCodeIntelSupportFields
            }
        }
    }

    fragment PreciseSupportFields on PreciseCodeIntelSupport {
        supportLevel
        indexers {
            ...CodeIntelIndexerFields
        }
    }

    fragment SearchBasedCodeIntelSupportFields on SearchBasedCodeIntelSupport {
        language
        supportLevel
    }

    fragment CodeIntelIndexerFields on CodeIntelIndexer {
        name
        url
    }

    ${preciseIndexFieldsFragment}
`
