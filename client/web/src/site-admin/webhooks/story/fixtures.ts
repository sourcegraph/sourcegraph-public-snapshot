import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import type {
    WebhookLogFields,
    WebhookLogPageHeaderExternalService,
    WebhookLogPageHeaderResult,
} from '../../../graphql-operations'
import { WEBHOOK_LOG_PAGE_HEADER } from '../backend'

export const BODY_JSON = '{"this is":"valid JSON","that should be":["re","indented"]}'
export const LARGE_BODY_JSON =
    '{"message": "webhooks: Add database Create method (#42639)\\n\\nAdd the Create method for the WebhookStore.\\r\\n\\r\\nThis change also switches to use two ids:\\r\\n\\r\\n`id` which is an auto-incremented id which we can use for sorting and pagination.\\r\\n`rand_id` which is an UUID and will be the user facing id use in the GraphQL layer\\r\\nand as part of the webhook URL.\\r\\n\\r\\nCo-authored-by: Someone"}'
export const BODY_PLAIN = 'this is definitely not valid JSON\n\tand should not be reformatted in any way'
export const LARGE_BODY_PLAIN = 'External service not found because we decided to be amazing and let you look for it'

export const HEADERS_JSON = [
    {
        name: 'Content-Type',
        values: ['application/json; charset=utf-8'],
    },
    {
        name: 'Content-Length',
        values: [BODY_JSON.length.toString()],
    },
    {
        name: 'X-Complex-Header',
        values: ['value 1', 'value 2'],
    },
]

export const LARGE_HEADERS_JSON = [
    {
        name: 'Content-Type',
        values: ['application/json; charset=utf-8'],
    },
    {
        name: 'Content-Length',
        values: [LARGE_BODY_JSON.length.toString()],
    },
    {
        name: 'X-Complex-Header',
        values: ['value 1', 'value 2'],
    },
    {
        name: 'X-Scheme',
        values: ['https'],
    },
]

export const HEADERS_PLAIN = [
    {
        name: 'Content-Type',
        values: ['text/plain'],
    },
    {
        name: 'Content-Length',
        values: [BODY_PLAIN.length.toString()],
    },
    {
        name: 'X-Complex-Header',
        values: ['value 1', 'value 2'],
    },
]

export const LARGE_HEADERS_PLAIN = [
    {
        name: 'Content-Type',
        values: ['text/plain'],
    },
    {
        name: 'Content-Length',
        values: [LARGE_BODY_PLAIN.length.toString()],
    },
    {
        name: 'X-Complex-Header',
        values: ['value 1', 'value 2'],
    },
]

export const webhookLogNode = (
    defaultArguments: Pick<WebhookLogFields, 'receivedAt' | 'statusCode'>,
    overrides?: Partial<WebhookLogFields>
): WebhookLogFields => ({
    id: overrides?.id ?? 'ID',
    receivedAt: overrides?.receivedAt ?? defaultArguments.receivedAt,
    statusCode: overrides?.statusCode ?? defaultArguments.statusCode,
    externalService: overrides?.externalService ?? null,
    request: overrides?.request ?? {
        headers: HEADERS_JSON,
        body: BODY_JSON,
        method: 'POST',
        url: '/my/url',
        version: 'HTTP/1.1',
    },
    response: overrides?.response ?? {
        headers: HEADERS_PLAIN,
        body: BODY_PLAIN,
    },
})

export const buildExternalServices = (count: number): WebhookLogPageHeaderExternalService[] => {
    const services: WebhookLogPageHeaderExternalService[] = []

    for (let index = 0; index < count; index++) {
        const name = `External service ${index}`
        services.push({
            __typename: 'ExternalService',
            id: name,
            displayName: name,
        })
    }

    return services
}

export const buildHeaderMock = (
    externalServiceCount: number,
    webhookLogCount: number
): MockedResponse<WebhookLogPageHeaderResult>[] => [
    {
        request: { query: getDocumentNode(WEBHOOK_LOG_PAGE_HEADER) },
        result: {
            data: {
                externalServices: {
                    totalCount: externalServiceCount,
                    nodes: buildExternalServices(externalServiceCount),
                },
                webhookLogs: {
                    totalCount: webhookLogCount,
                },
            },
        },
    },
]
