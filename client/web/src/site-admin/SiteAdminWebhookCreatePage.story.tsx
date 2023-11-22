import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { EXTERNAL_SERVICES } from '../components/externalServices/backend'
import { WebStory } from '../components/WebStory'

import { createExternalService } from './fixtures'
import { SiteAdminWebhookCreatePage } from './SiteAdminWebhookCreatePage'

const decorator: Decorator = Story => <Story />

const config: Meta = {
    title: 'web/site-admin/webhooks/incoming/SiteAdminWebhookCreatePage',
    decorators: [decorator],
}

export default config

export const WebhookCreatePage: StoryFn = () => {
    const mocks = new WildcardMockLink([
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
                            createExternalService(ExternalServiceKind.BITBUCKETCLOUD, 'https://bitbucket.org'),
                            createExternalService(ExternalServiceKind.BITBUCKETSERVER, 'https://sgdev.bitbucket.org'),
                            createExternalService(ExternalServiceKind.BITBUCKETSERVER, 'https://sgprod.bitbucket.org'),
                            createExternalService(ExternalServiceKind.GITLAB, 'https://gitlab.com'),
                            createExternalService(ExternalServiceKind.PERFORCE, 'https://perforce.com'),
                        ],
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])
    return (
        <WebStory>
            {() => (
                <MockedTestProvider link={mocks}>
                    <SiteAdminWebhookCreatePage telemetryService={NOOP_TELEMETRY_SERVICE} />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

WebhookCreatePage.storyName = 'Create webhook'

export const WebhookCreatePageWithError: StoryFn = () => {
    const mockedResponse: MockedResponse[] = [
        {
            request: {
                query: getDocumentNode(EXTERNAL_SERVICES),
                variables: { first: null, after: null, repo: null },
            },
            error: new Error('oops'),
        },
    ]
    return (
        <WebStory>
            {() => (
                <MockedTestProvider mocks={mockedResponse}>
                    <SiteAdminWebhookCreatePage telemetryService={NOOP_TELEMETRY_SERVICE} />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

WebhookCreatePageWithError.storyName = 'Error during external services fetch'
