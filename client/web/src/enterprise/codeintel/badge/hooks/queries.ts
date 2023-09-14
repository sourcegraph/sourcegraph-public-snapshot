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
