import { gql } from '@sourcegraph/http-client'

export const lsifUploadAuditLogsFieldsFragment = gql`
    fragment LsifUploadsAuditLogsFields on LSIFUploadAuditLog {
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

export const lsifUploadFieldsFragment = gql`
    fragment LsifUploadFields on LSIFUpload {
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
        state
        failure
        isLatestForRepo
        uploadedAt
        startedAt
        finishedAt
        placeInQueue
        associatedIndex {
            id
            state
            queuedAt
            startedAt
            finishedAt
            placeInQueue
        }
        auditLogs {
            ...LsifUploadsAuditLogsFields
        }
    }

    ${lsifUploadAuditLogsFieldsFragment}
`

export const lsifUploadConnectionFieldsFragment = gql`
    fragment LsifUploadConnectionFields on LSIFUploadConnection {
        nodes {
            ...LsifUploadFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }

    ${lsifUploadFieldsFragment}
`
