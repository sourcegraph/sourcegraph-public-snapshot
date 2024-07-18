import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    OutboundWebhookLogFields,
    OutboundWebhookLogsResult,
    OutboundWebhookLogsVariables,
} from '../../../graphql-operations'
import { WEBHOOK_LOG_REQUEST_FIELDS_FRAGMENT, WEBHOOK_LOG_RESPONSE_FIELDS_FRAGMENT } from '../../webhooks/backend'

const OUTBOUND_WEBHOOK_LOG_FIELDS_FRAGMENT = gql`
    ${WEBHOOK_LOG_REQUEST_FIELDS_FRAGMENT}
    ${WEBHOOK_LOG_RESPONSE_FIELDS_FRAGMENT}

    fragment OutboundWebhookLogFields on OutboundWebhookLog {
        id
        job {
            eventType
            payload
        }
        sentAt
        statusCode
        request {
            ...WebhookLogRequestFields
        }
        response {
            ...WebhookLogResponseFields
        }
        error
    }
`

export const OUTBOUND_WEBHOOK_LOGS = gql`
    ${OUTBOUND_WEBHOOK_LOG_FIELDS_FRAGMENT}

    query OutboundWebhookLogs($id: ID!, $onlyErrors: Boolean!, $first: Int, $after: String) {
        node(id: $id) {
            ... on OutboundWebhook {
                logs(first: $first, after: $after, onlyErrors: $onlyErrors) {
                    nodes {
                        ...OutboundWebhookLogFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        }
    }
`

export const useOutboundWebhookLogsConnection = (
    id: string,
    onlyErrors: boolean
): UseShowMorePaginationResult<OutboundWebhookLogsResult, OutboundWebhookLogFields> =>
    useShowMorePagination<OutboundWebhookLogsResult, OutboundWebhookLogsVariables, OutboundWebhookLogFields>({
        query: OUTBOUND_WEBHOOK_LOGS,
        variables: {
            id,
            onlyErrors,
        },
        getConnection: result => {
            const { node } = dataOrThrowErrors(result)
            if (node?.__typename !== 'OutboundWebhook') {
                throw new Error('unexpected node type')
            }
            return node.logs
        },
        options: { pollInterval: 5000 },
    })
