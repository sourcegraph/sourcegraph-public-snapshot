import { MockedResponse } from '@apollo/client/testing'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { EXTERNAL_SERVICES } from '../components/externalServices/backend'
import { WebStory } from '../components/WebStory'

import { createExternalService } from './fixtures'
import { SiteAdminWebhookCreatePage } from './SiteAdminWebhookCreatePage'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/SiteAdminWebhookCreatePage',
    decorators: [decorator],
}

export default config

export const WebhookCreatePage: Story = () => {
    const mocks = new WildcardMockLink([
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
                    <SiteAdminWebhookCreatePage
                        match={{} as any}
                        history={H.createMemoryHistory()}
                        location={{} as any}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

WebhookCreatePage.storyName = 'Create webhook'

export const WebhookCreatePageWithError: Story = () => {
    const mockedResponse: MockedResponse[] = [
        {
            request: {
                query: getDocumentNode(EXTERNAL_SERVICES),
                variables: { first: null, after: null },
            },
            error: new Error('oops'),
        },
    ]
    return (
        <WebStory>
            {() => (
                <MockedTestProvider mocks={mockedResponse}>
                    <SiteAdminWebhookCreatePage
                        match={{} as any}
                        history={H.createMemoryHistory()}
                        location={{} as any}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

WebhookCreatePageWithError.storyName = 'Error during external services fetch'
