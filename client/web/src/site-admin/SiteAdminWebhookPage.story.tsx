import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { addMinutes, formatRFC3339 } from 'date-fns'
import { Route, Routes } from 'react-router-dom'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'
import type { WebhookLogFields } from '../graphql-operations'

import { WEBHOOK_BY_ID } from './backend'
import { createWebhookMock, TIMESTAMP_MOCK } from './fixtures'
import { SiteAdminWebhookPage } from './SiteAdminWebhookPage'
import { WEBHOOK_BY_ID_LOG_PAGE_HEADER, WEBHOOK_LOGS_BY_ID } from './webhooks/backend'
import { BODY_JSON, BODY_PLAIN, HEADERS_JSON, HEADERS_PLAIN } from './webhooks/story/fixtures'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/site-admin/webhooks/incoming/SiteAdminWebhookPage',
    decorators: [decorator],
}

export default config

const WEBHOOK_MOCK_DATA = buildWebhookLogs()
const ERRORED_WEBHOOK_MOCK_DATA = WEBHOOK_MOCK_DATA.filter(webhook => webhook.statusCode !== 200)

export const SiteAdminWebhookPageStory: Story = args => {
    const buildWebhookLogsMock = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WEBHOOK_BY_ID),
                variables: {
                    id: '1',
                },
            },
            result: {
                data: {
                    node: createWebhookMock(ExternalServiceKind.GITHUB, 'https://github.com/'),
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WEBHOOK_LOGS_BY_ID),
                variables: {
                    first: 20,
                    after: null,
                    onlyErrors: false,
                    onlyUnmatched: false,
                    webhookID: '',
                },
            },
            result: {
                data: {
                    webhookLogs: {
                        nodes: WEBHOOK_MOCK_DATA,
                        pageInfo: { hasNextPage: false },
                        totalCount: 20,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WEBHOOK_LOGS_BY_ID),
                variables: {
                    first: 20,
                    after: null,
                    onlyErrors: true,
                    onlyUnmatched: false,
                    webhookID: '1',
                },
            },
            result: {
                data: {
                    webhookLogs: {
                        nodes: ERRORED_WEBHOOK_MOCK_DATA,
                        pageInfo: { hasNextPage: false },
                        totalCount: 20,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WEBHOOK_BY_ID_LOG_PAGE_HEADER),
                variables: {
                    webhookID: '1',
                },
            },
            result: {
                data: {
                    webhookLogs: {
                        totalCount: 13,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory initialEntries={['/site-admin/webhooks/incoming/1']}>
            {() => (
                <MockedTestProvider link={buildWebhookLogsMock}>
                    <Routes>
                        <Route
                            path="/site-admin/webhooks/incoming/:id"
                            element={
                                <SiteAdminWebhookPage
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                    telemetryRecorder={noOpTelemetryRecorder}
                                />
                            }
                        />
                    </Routes>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

SiteAdminWebhookPageStory.storyName = 'Incoming webhook'

export const SiteAdminWebhookPageWithoutLogsStory: Story = args => {
    const buildWebhookLogsMock = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WEBHOOK_BY_ID),
                variables: {
                    id: '',
                },
            },
            result: {
                data: {
                    node: createWebhookMock(ExternalServiceKind.GITHUB, 'https://github.com/'),
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WEBHOOK_LOGS_BY_ID),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: {
                    webhookLogs: {
                        nodes: [],
                        pageInfo: { hasNextPage: false },
                        totalCount: 0,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WEBHOOK_BY_ID_LOG_PAGE_HEADER),
                variables: {
                    webhookID: '',
                },
            },
            result: {
                data: {
                    webhookLogs: {
                        totalCount: 0,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {() => (
                <MockedTestProvider link={buildWebhookLogsMock}>
                    <SiteAdminWebhookPage
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

SiteAdminWebhookPageWithoutLogsStory.storyName = 'Incoming webhook without logs'

function buildWebhookLogs(): WebhookLogFields[] {
    const logs: WebhookLogFields[] = []

    for (let index = 0; index < 20; index++) {
        const externalServiceID = index % 5
        const statusCode = index % 3 === 0 ? 200 : index % 3 === 1 ? 400 : 500

        logs.push({
            __typename: 'WebhookLog',
            id: index.toString(),
            receivedAt: formatRFC3339(addMinutes(TIMESTAMP_MOCK, index)),
            externalService: {
                displayName: `External service ${externalServiceID}`,
            },
            statusCode,
            request: {
                headers: HEADERS_JSON,
                body: BODY_JSON,
                method: 'POST',
                url: '/my/url',
                version: 'HTTP/1.1',
            },
            response: {
                headers: HEADERS_PLAIN,
                body: BODY_PLAIN,
            },
        })
    }

    return logs
}
