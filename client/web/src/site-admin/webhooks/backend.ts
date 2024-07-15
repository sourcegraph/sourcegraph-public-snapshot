import { gql } from '@sourcegraph/http-client'

export const WEBHOOK_LOG_REQUEST_FIELDS_FRAGMENT = gql`
    fragment WebhookLogRequestFields on WebhookLogRequest {
        headers {
            name
            values
        }
        body
        method
        url
        version
    }
`

export const WEBHOOK_LOG_RESPONSE_FIELDS_FRAGMENT = gql`
    fragment WebhookLogResponseFields on WebhookLogResponse {
        headers {
            name
            values
        }
        body
    }
`

const WEBHOOK_LOG_FIELDS_FRAGMENT = gql`
    ${WEBHOOK_LOG_REQUEST_FIELDS_FRAGMENT}
    ${WEBHOOK_LOG_RESPONSE_FIELDS_FRAGMENT}

    fragment WebhookLogFields on WebhookLog {
        id
        receivedAt
        externalService {
            displayName
        }
        statusCode
        request {
            ...WebhookLogRequestFields
        }
        response {
            ...WebhookLogResponseFields
        }
    }
`

export const WEBHOOK_LOG_PAGE_HEADER = gql`
    query WebhookLogPageHeader {
        externalServices {
            nodes {
                ...WebhookLogPageHeaderExternalService
            }
            totalCount
        }

        webhookLogs(onlyErrors: true, legacyOnly: true) {
            totalCount
        }
    }

    fragment WebhookLogPageHeaderExternalService on ExternalService {
        id
        displayName
    }
`

export const WEBHOOK_LOGS_BY_ID = gql`
    ${WEBHOOK_LOG_FIELDS_FRAGMENT}

    query WebhookLogsByWebhookID(
        $first: Int
        $after: String
        $onlyErrors: Boolean!
        $onlyUnmatched: Boolean!
        $webhookID: ID!
    ) {
        webhookLogs(
            first: $first
            after: $after
            onlyErrors: $onlyErrors
            onlyUnmatched: $onlyUnmatched
            webhookID: $webhookID
        ) {
            ...ListWebhookLogs
        }
    }

    fragment ListWebhookLogs on WebhookLogConnection {
        nodes {
            ...WebhookLogFields
        }
        pageInfo {
            hasNextPage
            endCursor
        }
        totalCount
    }
`

export const WEBHOOK_BY_ID_LOG_PAGE_HEADER = gql`
    query WebhookByIDLogPageHeader($webhookID: ID!) {
        webhookLogs(webhookID: $webhookID, onlyErrors: true) {
            totalCount
        }
    }
`
