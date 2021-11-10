import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { Scalars } from '../../graphql-operations'

export type SelectedExternalService = 'unmatched' | 'all' | Scalars['ID']

export const WEBHOOK_LOG_CONNECTION_FIELDS_FRAGMENT = gql`
    fragment WebhookLogConnectionFields on WebhookLogConnection {
        nodes {
            __typename
            ...WebhookLogFields
        }
        pageInfo {
            hasNextPage
            endCursor
        }
        totalCount
    }

    fragment WebhookLogFields on WebhookLog {
        id
        receivedAt
        externalService {
            displayName
        }
        statusCode
        request {
            __typename
            ...WebhookLogMessageFields
            ...WebhookLogRequestFields
        }
        response {
            __typename
            ...WebhookLogMessageFields
        }
    }

    fragment WebhookLogMessageFields on WebhookLogMessage {
        headers {
            name
            values
        }
        body
    }

    fragment WebhookLogRequestFields on WebhookLogRequest {
        method
        url
        version
    }
`

export const GLOBAL_WEBHOOK_LOGS = gql`
    query WebhookLogs($first: Int, $after: String, $onlyErrors: Boolean!, $onlyUnmatched: Boolean!) {
        webhookLogs(first: $first, after: $after, onlyErrors: $onlyErrors, onlyUnmatched: $onlyUnmatched) {
            ...WebhookLogConnectionFields
        }
    }

    ${WEBHOOK_LOG_CONNECTION_FIELDS_FRAGMENT}
`

export const EXTERNAL_SERVICE_WEBHOOK_LOGS = gql`
    query ServiceWebhookLogs($first: Int, $after: String, $id: ID!, $onlyErrors: Boolean!) {
        node(id: $id) {
            ... on ExternalService {
                __typename
                webhookLogs(first: $first, after: $after, onlyErrors: $onlyErrors) {
                    ...WebhookLogConnectionFields
                }
            }
        }
    }

    ${WEBHOOK_LOG_CONNECTION_FIELDS_FRAGMENT}
`

export const WEBHOOK_LOG_PAGE_HEADER = gql`
    query WebhookLogPageHeader {
        externalServices {
            nodes {
                ...WebhookLogPageHeaderExternalService
            }
            totalCount
        }

        webhookLogs(onlyErrors: true) {
            totalCount
        }
    }

    fragment WebhookLogPageHeaderExternalService on ExternalService {
        id
        displayName
    }
`
