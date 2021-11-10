import { storiesOf } from '@storybook/react'
import { addMinutes, formatRFC3339 } from 'date-fns'
import React from 'react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/shared/src/graphql/apollo'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { ServiceWebhookLogsVariables, WebhookLogFields, WebhookLogsVariables } from '../../graphql-operations'

import { EXTERNAL_SERVICE_WEBHOOK_LOGS, GLOBAL_WEBHOOK_LOGS } from './backend'
import { BODY_JSON, BODY_PLAIN, buildHeaderMock, HEADERS_JSON, HEADERS_PLAIN } from './story/fixtures'
import { WebhookLogPage } from './WebhookLogPage'

const { add } = storiesOf('web/site-admin/webhooks/WebhookLogPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const buildWebhookLogs = (count: number, externalServiceCount: number): WebhookLogFields[] => {
    const logs: WebhookLogFields[] = []
    const time = new Date(2021, 10, 8, 16, 40, 30)

    for (let index = 0; index < count; index++) {
        const externalServiceID = index % (externalServiceCount + 1)
        const statusCode =
            index % 3 === 0
                ? 200 + Math.floor(index / 3)
                : index % 3 === 1
                ? 400 + Math.floor(index / 3)
                : 500 + Math.floor(index / 3)

        logs.push({
            __typename: 'WebhookLog',
            id: index.toString(),
            receivedAt: formatRFC3339(addMinutes(time, index)),
            externalService:
                externalServiceID === externalServiceCount
                    ? null
                    : {
                          displayName: `External service ${externalServiceID}`,
                      },
            statusCode,
            request: {
                __typename: 'WebhookLogRequest',
                headers: HEADERS_JSON.map(header => ({ ...header, __typename: 'WebhookLogHeader' })),
                body: BODY_JSON,
                method: 'POST',
                url: '/my/url',
                version: 'HTTP/1.1',
            },
            response: {
                __typename: 'WebhookLogResponse',
                headers: HEADERS_PLAIN.map(header => ({ ...header, __typename: 'WebhookLogHeader' })),
                body: BODY_PLAIN,
            },
        })
    }

    return logs
}

const buildWebhookLogMockLink = (count: number, externalServiceCount: number): WildcardMockLink => {
    const logs = buildWebhookLogs(count, externalServiceCount)

    return new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(GLOBAL_WEBHOOK_LOGS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: variables => {
                let { first, after, onlyErrors, onlyUnmatched } = variables as WebhookLogsVariables

                const filtered = logs.filter(log => {
                    if (onlyErrors && log.statusCode < 400) {
                        return false
                    }
                    if (onlyUnmatched && log.externalService) {
                        return false
                    }

                    return true
                })

                first = first ?? 20
                const afterNumber = after?.length ? +after : 0
                const page = filtered.slice(afterNumber, afterNumber + first)
                const cursor = afterNumber + first

                return {
                    data: {
                        webhookLogs: {
                            __typename: 'WebhookLogConnection',
                            nodes: page,
                            pageInfo: {
                                __typename: 'PageInfo',
                                hasNextPage: logs.length > cursor,
                                endCursor: cursor.toString(),
                            },
                            totalCount: logs.length,
                        },
                    },
                }
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(EXTERNAL_SERVICE_WEBHOOK_LOGS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: variables => {
                let { first, after, onlyErrors, id } = variables as ServiceWebhookLogsVariables

                const filtered = logs.filter(log => {
                    if (onlyErrors && log.statusCode < 400) {
                        return false
                    }

                    return id === log.externalService?.displayName
                })

                first = first ?? 20
                const afterNumber = after?.length ? +after : 0
                const page = filtered.slice(afterNumber, afterNumber + first)
                const cursor = afterNumber + first

                return {
                    data: {
                        node: {
                            __typename: 'ExternalService',
                            webhookLogs: {
                                __typename: 'WebhookLogConnection',
                                nodes: page,
                                pageInfo: {
                                    __typename: 'PageInfo',
                                    hasNextPage: filtered.length > cursor,
                                    endCursor: cursor.toString(),
                                },
                                totalCount: filtered.length,
                            },
                        },
                    },
                }
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        ...buildHeaderMock(externalServiceCount, Math.floor((count * 2) / 3)),
    ])
}

add('no logs', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildWebhookLogMockLink(0, 2)}>
                <WebhookLogPage {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('one page of logs', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildWebhookLogMockLink(20, 2)}>
                <WebhookLogPage {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('two pages of logs', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildWebhookLogMockLink(40, 2)}>
                <WebhookLogPage {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
))
