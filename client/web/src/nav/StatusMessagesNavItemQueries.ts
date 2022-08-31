import { gql } from '@sourcegraph/http-client'

export const STATUS_MESSAGES = gql`
    query StatusMessages {
        statusMessages {
            ...StatusMessageFields
        }
    }

    fragment StatusMessageFields on StatusMessage {
        type: __typename

        ... on CloningProgress {
            message
        }

        ... on SyncError {
            message
        }

        ... on ExternalServiceSyncError {
            message
            externalService {
                id
                displayName
            }
        }
    }
`
