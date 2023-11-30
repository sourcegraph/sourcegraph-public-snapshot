import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { Route, Routes } from 'react-router-dom'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { EXTERNAL_SERVICES } from '../components/externalServices/backend'
import { WebStory } from '../components/WebStory'

import { WEBHOOK_BY_ID } from './backend'
import { createExternalService, createWebhookMock } from './fixtures'
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
                                query: getDocumentNode(EXTERNAL_SERVICES),
                                variables: { first: null, after: null, repo: null },
                            },
                            result: {
                                data: {
                                    externalServices: {
                                        __typename: 'ExternalServiceConnection',
                                        totalCount: 17,
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
                        element={<SiteAdminWebhookUpdatePage telemetryService={NOOP_TELEMETRY_SERVICE} />}
                    />
                </Routes>
            </MockedTestProvider>
        )}
    </WebStory>
)

WebhookUpdatePage.storyName = 'Update webhook'
