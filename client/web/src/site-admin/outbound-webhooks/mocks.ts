import type { FetchResult } from '@apollo/client'
import type { MockedResponse } from '@apollo/client/testing'
import { type GraphQLRequestWithWildcard, MATCH_ANY_PARAMETERS, type WildcardMockedResponse } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'

import type {
    OutboundWebhookByIDResult,
    OutboundWebhookEventTypesResult,
    OutboundWebhookFieldsWithStats,
    OutboundWebhookLogFields,
    OutboundWebhookLogsResult,
    OutboundWebhooksListResult,
    WebhookLogResponseFields,
} from '../../graphql-operations'

import { OUTBOUND_WEBHOOKS, OUTBOUND_WEBHOOK_BY_ID, OUTBOUND_WEBHOOK_EVENT_TYPES } from './backend'
import { OUTBOUND_WEBHOOK_LOGS } from './logs/backend'

export interface WildcardResponse<TData> extends Omit<Omit<WildcardMockedResponse, 'result'>, 'response'> {
    request: GraphQLRequestWithWildcard
    result: FetchResult<TData>
}

export const eventTypesMock: MockedResponse<OutboundWebhookEventTypesResult> = {
    request: {
        query: getDocumentNode(OUTBOUND_WEBHOOK_EVENT_TYPES),
    },
    result: {
        data: {
            outboundWebhookEventTypes: [
                {
                    __typename: 'OutboundWebhookEventType',
                    key: 'batch_change:apply',
                    description: 'sent when a batch change is applied',
                },
                {
                    __typename: 'OutboundWebhookEventType',
                    key: 'batch_change:close',
                    description: 'sent when a batch change is closed',
                },
            ],
        },
    },
}

export const buildOutboundWebhookMock = (id: string): MockedResponse<OutboundWebhookByIDResult> => ({
    request: { query: getDocumentNode(OUTBOUND_WEBHOOK_BY_ID), variables: { id } },
    result: {
        data: {
            node: {
                __typename: 'OutboundWebhook',
                id,
                url: 'http://example.com/',
                eventTypes: [{ eventType: 'batch_change:apply', scope: null }],
            },
        },
    },
})

enum LogState {
    OK,
    ServerError,
    NetworkError,
}

const randomLog = (state: LogState): OutboundWebhookLogFields => {
    const payload = '{"a vaguely": "plausible", "webhook": "payload"}'

    let statusCode = 200
    let response: WebhookLogResponseFields | null = null
    let error: string | null = null
    if (state === LogState.ServerError) {
        statusCode = 500
        response = {
            __typename: 'WebhookLogResponse',
            headers: [{ name: 'content-type', values: ['application/json'] }],
            body: '"success"',
        }
    } else if (state === LogState.NetworkError) {
        statusCode = 0
        error = 'Network error'
    }

    return {
        __typename: 'OutboundWebhookLog',
        id: Math.floor(Math.random() * 1_000_000).toLocaleString(),
        job: {
            __typename: 'OutboundWebhookJob',
            eventType: 'batch_change:apply',
            payload,
        },
        sentAt: '2022-12-30T13:03:00Z',
        statusCode,
        request: {
            __typename: 'WebhookLogRequest',
            headers: [
                { name: 'content-type', values: ['application/json'] },
                { name: 'x-sourcegraph-webhook-signature', values: ['abcdef'] },
            ],
            body: payload,
            method: 'POST',
            url: 'http://example.com/',
            version: '',
        },
        response,
        error,
    }
}

const randomLogs = (count: number): OutboundWebhookLogFields[] => {
    const logs = []
    for (let idx = 0; idx < count; idx++) {
        logs.push(
            randomLog(idx % 10 === 0 ? LogState.NetworkError : idx % 10 === 1 ? LogState.ServerError : LogState.OK)
        )
    }

    return logs
}

export const logConnectionLink: WildcardResponse<OutboundWebhookLogsResult> = {
    request: {
        query: getDocumentNode(OUTBOUND_WEBHOOK_LOGS),
        variables: MATCH_ANY_PARAMETERS,
    },
    result: {
        data: {
            node: {
                __typename: 'OutboundWebhook',
                logs: {
                    nodes: randomLogs(20),
                    totalCount: 50,
                    pageInfo: { hasNextPage: false },
                },
            },
        },
    },
    nMatches: Number.POSITIVE_INFINITY,
}

const randomWebhook = (num: number): OutboundWebhookFieldsWithStats => ({
    __typename: 'OutboundWebhook',
    id: `${num}`,
    url: `http://example.com/${num}`,
    eventTypes: [{ eventType: 'batch_change:apply', scope: null }],
    stats: { total: num * 10, errored: Math.floor(num * 0.5) },
})

const randomWebhooks = (count: number): OutboundWebhookFieldsWithStats[] => {
    const webhooks = []
    for (let idx = 0; idx < count; idx++) {
        webhooks.push(randomWebhook(idx))
    }

    return webhooks
}

export const buildOutboundWebhooksConnectionLink = (count: number): WildcardResponse<OutboundWebhooksListResult> => ({
    request: {
        query: getDocumentNode(OUTBOUND_WEBHOOKS),
        variables: MATCH_ANY_PARAMETERS,
    },
    result: {
        data: {
            outboundWebhooks: {
                __typename: 'OutboundWebhookConnection',
                nodes: randomWebhooks(count),
                totalCount: count,
                pageInfo: { hasNextPage: false },
            },
        },
    },
    nMatches: Number.POSITIVE_INFINITY,
})
