import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { Route, Routes } from 'react-router-dom'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'
import { WebhookExternalServiceFields } from '../graphql-operations'

import { WEBHOOK_BY_ID, WEBHOOK_EXTERNAL_SERVICES } from './backend'
import { createWebhookMock } from './fixtures'
import { SiteAdminWebhookUpdatePage } from './SiteAdminWebhookUpdatePage'

const decorator: Decorator = Story => <Story />

const config: Meta = {
    title: 'web/site-admin/webhooks/incoming/SiteAdminWebhookUpdatePage',
    decorators: [decorator],
}

export default config

export const WebhookUpdatePage: StoryFn = () => (
    <WebStory initialEntries={['/site-admin/webhooks/incoming/1']}>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(WEBHOOK_EXTERNAL_SERVICES),
                                variables: {},
                            },
                            result: {
                                data: {
                                    externalServices: {
                                        __typename: 'ExternalServiceConnection',
                                        totalCount: 6,
                                        pageInfo: {
                                            endCursor: null,
                                            hasNextPage: false,
                                        },
                                        nodes: [
                                            createExternalService(ExternalServiceKind.GITHUB, 'https://github.com'),
                                            createExternalService(
                                                ExternalServiceKind.BITBUCKETCLOUD,
                                                'https://bitbucket.org'
                                            ),
                                            createExternalService(
                                                ExternalServiceKind.BITBUCKETSERVER,
                                                'https://sgdev.bitbucket.org'
                                            ),
                                            createExternalService(
                                                ExternalServiceKind.BITBUCKETSERVER,
                                                'https://sgprod.bitbucket.org'
                                            ),
                                            createExternalService(ExternalServiceKind.GITLAB, 'https://gitlab.com'),
                                            createExternalService(ExternalServiceKind.PERFORCE, 'https://perforce.com'),
                                        ],
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                        {
                            request: {
                                query: getDocumentNode(WEBHOOK_BY_ID),
                                variables: {
                                    id: '1',
                                },
                            },
                            result: {
                                data: {
                                    node: createWebhookMock(
                                        ExternalServiceKind.BITBUCKETSERVER,
                                        'https://sgprod.bitbucket.org'
                                    ),
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <Routes>
                    <Route
                        path="/site-admin/webhooks/incoming/:id"
                        element={
                            <div className="container p-4">
                                <SiteAdminWebhookUpdatePage
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                    telemetryRecorder={noOpTelemetryRecorder}
                                />
                            </div>
                        }
                    />
                </Routes>
            </MockedTestProvider>
        )}
    </WebStory>
)

WebhookUpdatePage.storyName = 'Update webhook'

function createExternalService(kind: ExternalServiceKind, url: string): WebhookExternalServiceFields {
    return {
        __typename: 'ExternalService',
        id: `service-${url}`,
        kind,
        displayName: `${kind}-123`,
        url,
    }
}
