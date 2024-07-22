import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { type ExternalServiceFields, ExternalServiceKind } from '../../graphql-operations'
import { WebStory, type WebStoryChildrenProps } from '../WebStory'

import { FETCH_EXTERNAL_SERVICE } from './backend'
import { ExternalServiceEditPage } from './ExternalServiceEditPage'

const decorator: Decorator = story => (
    <div className="p-3 container">
        <WebStory
            path="/site-admin/external-services/:externalServiceID/edit"
            initialEntries={['/site-admin/external-services/service123/edit']}
        >
            {story}
        </WebStory>
    </div>
)

const config: Meta = {
    title: 'web/External services/ExternalServiceEditPage',
    parameters: {},
    decorators: [decorator],
}

export default config

const externalService = {
    __typename: 'ExternalService',
    id: 'service123',
    kind: ExternalServiceKind.GITHUB,
    warning: null,
    config: '{"githubconfig": true}',
    displayName: 'GitHub.com',
    webhookURL: null,
    lastSyncError: null,
    repoCount: 0,
    lastSyncAt: null,
    unrestricted: false,
    nextSyncAt: null,
    updatedAt: '2021-03-15T19:39:11Z',
    createdAt: '2021-03-15T19:39:11Z',
    hasConnectionCheck: false,
    namespace: {
        id: 'userid',
        namespaceName: 'johndoe',
        url: '/users/johndoe',
    },
    rateLimiterState: {
        __typename: 'RateLimiterState',
        burst: 10,
        currentCapacity: 10,
        infinite: false,
        interval: 3600,
        lastReplenishment: new Date().toISOString(),
        limit: 5,
    },
    creator: {
        __typename: 'User',
        username: 'alice',
        url: '/users/alice',
    },
    lastUpdater: {
        __typename: 'User',
        username: 'alice',
        url: '/users/alice',
    },
} as ExternalServiceFields

function newFetchMock(node: ExternalServiceFields): WildcardMockLink {
    return new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(FETCH_EXTERNAL_SERVICE),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: { node } },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])
}

export const ViewConfig: StoryFn<WebStoryChildrenProps> = props => (
    <MockedTestProvider link={newFetchMock(externalService)}>
        <ExternalServiceEditPage
            telemetryService={NOOP_TELEMETRY_SERVICE}
            telemetryRecorder={noOpTelemetryRecorder}
            autoFocusForm={false}
            externalServicesFromFile={false}
            allowEditExternalServicesWithFile={false}
        />
    </MockedTestProvider>
)

ViewConfig.storyName = 'View external service config'

export const ConfigWithInvalidUrl: StoryFn<WebStoryChildrenProps> = props => (
    <MockedTestProvider link={newFetchMock({ ...externalService, config: '{"url": "invalid-url"}' })}>
        <ExternalServiceEditPage
            telemetryService={NOOP_TELEMETRY_SERVICE}
            telemetryRecorder={noOpTelemetryRecorder}
            autoFocusForm={false}
            externalServicesFromFile={false}
            allowEditExternalServicesWithFile={false}
        />
    </MockedTestProvider>
)

ConfigWithInvalidUrl.storyName = 'External service config with invalid url'

export const ConfigWithWarning: StoryFn<WebStoryChildrenProps> = props => (
    <MockedTestProvider link={newFetchMock({ ...externalService, warning: 'Invalid config we could not sync stuff' })}>
        <ExternalServiceEditPage
            telemetryService={NOOP_TELEMETRY_SERVICE}
            telemetryRecorder={noOpTelemetryRecorder}
            autoFocusForm={false}
            externalServicesFromFile={false}
            allowEditExternalServicesWithFile={false}
        />
    </MockedTestProvider>
)

ConfigWithWarning.storyName = 'External service config with warning after update'

export const EditingDisabled: StoryFn<WebStoryChildrenProps> = props => (
    <MockedTestProvider link={newFetchMock({ ...externalService, warning: 'Invalid config we could not sync stuff' })}>
        <ExternalServiceEditPage
            telemetryService={NOOP_TELEMETRY_SERVICE}
            telemetryRecorder={noOpTelemetryRecorder}
            autoFocusForm={false}
            externalServicesFromFile={true}
            allowEditExternalServicesWithFile={false}
        />
    </MockedTestProvider>
)

EditingDisabled.storyName = 'External service config EXTSVC_CONFIG_FIlE set'
