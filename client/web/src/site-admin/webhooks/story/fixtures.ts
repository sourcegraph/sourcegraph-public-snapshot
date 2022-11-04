import { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import {
    WebhookLogFields,
    WebhookLogPageHeaderExternalService,
    WebhookLogPageHeaderResult,
} from '../../../graphql-operations'
import { WEBHOOK_LOG_PAGE_HEADER } from '../backend'

export const BODY_JSON = '{"this is":"valid JSON","that should be":["re","indented"]}'
export const BODY_PLAIN = 'this is definitely not valid JSON\n\tand should not be reformatted in any way'

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
