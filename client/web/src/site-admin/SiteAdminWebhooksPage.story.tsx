import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'
import type { WebhookFields } from '../graphql-operations'

import { WEBHOOKS, WEBHOOK_PAGE_HEADER } from './backend'
import { SiteAdminWebhooksPage } from './SiteAdminWebhooksPage'

const decorator: Decorator = Story => <Story />

const config: Meta = {
    title: 'web/site-admin/webhooks/incoming/SiteAdminWebhooksPage',
    decorators: [decorator],
}

export default config

export const NoWebhooksFound: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(WEBHOOKS),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    webhooks: {
                                        nodes: [],
                                        totalCount: 0,
                                        pageInfo: {
                                            hasNextPage: false,
                                        },
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                        {
                            request: {
                                query: getDocumentNode(WEBHOOK_PAGE_HEADER),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    webhooks: {
                                        nodes: [],
                                    },
                                    errorsOnly: {
                                        nodes: [],
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <div className="container p-4">
                    <SiteAdminWebhooksPage
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </div>
            </MockedTestProvider>
        )}
    </WebStory>
)

NoWebhooksFound.storyName = 'No webhooks found'

export const FiveWebhooksFound: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(WEBHOOKS),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    webhooks: {
                                        nodes: [
                                            createWebhookMock(
                                                'Bitbucket Cloud commit webhook',
                                                ExternalServiceKind.BITBUCKETCLOUD,
                                                'https://bitbucket.com/'
                                            ),
                                            createWebhookMock(
                                                'Github.com commit webhook',
                                                ExternalServiceKind.GITHUB,
                                                'https://github.com/'
                                            ),
                                            createWebhookMock(
                                                'Github.com PR push webhook',
                                                ExternalServiceKind.GITHUB,
                                                'https://github.com/'
                                            ),
                                            createWebhookMock(
                                                'Github.com PR creation webhook',
                                                ExternalServiceKind.GITHUB,
                                                'https://github.com/'
                                            ),
                                            createWebhookMock(
                                                'Bitbucket Cloud PR webhook',
                                                ExternalServiceKind.BITBUCKETCLOUD,
                                                'https://bitbucket.com/'
                                            ),
                                        ],
                                        totalCount: 5,
                                        pageInfo: {
                                            hasNextPage: false,
                                        },
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                        {
                            request: {
                                query: getDocumentNode(WEBHOOK_PAGE_HEADER),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    webhooks: {
                                        nodes: [
                                            {
                                                webhookLogs: {
                                                    totalCount: 2,
                                                },
                                            },
                                            {
                                                webhookLogs: {
                                                    totalCount: 0,
                                                },
                                            },
                                            {
                                                webhookLogs: {
                                                    totalCount: 1,
                                                },
                                            },
                                            {
                                                webhookLogs: {
                                                    totalCount: 0,
                                                },
                                            },
                                            {
                                                webhookLogs: {
                                                    totalCount: 2,
                                                },
                                            },
                                        ],
                                    },
                                    errorsOnly: {
                                        nodes: [],
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <div className="container p-4">
                    <SiteAdminWebhooksPage
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </div>
            </MockedTestProvider>
        )}
    </WebStory>
)

FiveWebhooksFound.storyName = '5 webhooks found'

function createWebhookMock(name: string, kind: ExternalServiceKind, urn: string): WebhookFields {
    return {
        __typename: 'Webhook',
        createdAt: '',
        id: `webhook-${urn}`,
        name,
        secret: null,
        updatedAt: '',
        url: '',
        uuid: '',
        codeHostKind: kind,
        codeHostURN: urn,
        createdBy: {
            username: 'alice',
            url: 'users/alice',
        },
        updatedBy: {
            username: 'alice',
            url: 'users/alice',
        },
    }
}
