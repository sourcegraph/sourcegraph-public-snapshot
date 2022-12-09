import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'
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

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/SiteAdminWebhookUpdatePage',
    decorators: [decorator],
}

export default config

export const WebhookUpdatePage: Story = args => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(EXTERNAL_SERVICES),
                                variables: { first: null, after: null },
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
                <SiteAdminWebhookUpdatePage
                    match={args.match}
                    history={H.createMemoryHistory()}
                    location={{} as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

WebhookUpdatePage.storyName = 'Update webhook'
WebhookUpdatePage.args = {
    match: {
        params: {
            id: '1',
        },
    },
}
