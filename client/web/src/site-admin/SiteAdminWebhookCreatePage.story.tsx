import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../components/WebStory'
import { WebhookExternalServiceFields } from '../graphql-operations'

import { WEBHOOK_EXTERNAL_SERVICES } from './backend'
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
                    <div className="container p-4">
                        <SiteAdminWebhookCreatePage
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </div>
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
                query: getDocumentNode(WEBHOOK_EXTERNAL_SERVICES),
                variables: {},
            },
            error: new Error('oops'),
        },
    ]
    return (
        <WebStory>
            {() => (
                <MockedTestProvider mocks={mockedResponse}>
                    <div className="container p-4">
                        <SiteAdminWebhookCreatePage
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </div>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

WebhookCreatePageWithError.storyName = 'Error during external services fetch'

function createExternalService(kind: ExternalServiceKind, url: string): WebhookExternalServiceFields {
    return {
        __typename: 'ExternalService',
        id: `service-${url}`,
        kind,
        displayName: `${kind}-123`,
        url,
    }
}
