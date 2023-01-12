import { DecoratorFn, Story, Meta } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { ExternalServiceFields, ExternalServiceKind } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { FETCH_EXTERNAL_SERVICE } from './backend'
import { ExternalServiceEditPage } from './ExternalServiceEditPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/External services/ExternalServiceEditPage',
    parameters: {
        chromatic: {
            // Delay screenshot taking, so Monaco has some time to get syntax highlighting prepared.
            delay: 2000,
        },
    },
    decorators: [decorator],
}

export default config

const externalService = {
    __typename: 'ExternalService' as const,
    id: 'service123',
    kind: ExternalServiceKind.GITHUB,
    warning: null,
    config: '{"githubconfig": true}',
    displayName: 'GitHub.com',
    webhookURL: null,
    lastSyncError: null,
    repoCount: 0,
    lastSyncAt: null,
    nextSyncAt: null,
    updatedAt: '2021-03-15T19:39:11Z',
    createdAt: '2021-03-15T19:39:11Z',
    hasConnectionCheck: false,
    namespace: {
        id: 'userid',
        namespaceName: 'johndoe',
        url: '/users/johndoe',
    },
}

function newFetchMock(node: { __typename: 'ExternalService' } & ExternalServiceFields): WildcardMockLink {
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

export const ViewConfig: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider link={newFetchMock(externalService)}>
                <ExternalServiceEditPage
                    {...webProps}
                    routingPrefix="/site-admin"
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    externalServiceID="service123"
                    autoFocusForm={false}
                    externalServicesFromFile={false}
                    allowEditExternalServicesWithFile={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ViewConfig.storyName = 'View external service config'

export const ConfigWithInvalidUrl: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider link={newFetchMock({ ...externalService, config: '{"url": "invalid-url"}' })}>
                <ExternalServiceEditPage
                    {...webProps}
                    routingPrefix="/site-admin"
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    externalServiceID="service123"
                    autoFocusForm={false}
                    externalServicesFromFile={false}
                    allowEditExternalServicesWithFile={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfigWithInvalidUrl.storyName = 'External service config with invalid url'

export const ConfigWithWarning: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider
                link={newFetchMock({ ...externalService, warning: 'Invalid config we could not sync stuff' })}
            >
                <ExternalServiceEditPage
                    {...webProps}
                    routingPrefix="/site-admin"
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    externalServiceID="service123"
                    autoFocusForm={false}
                    externalServicesFromFile={false}
                    allowEditExternalServicesWithFile={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfigWithWarning.storyName = 'External service config with warning after update'

export const EditingDisabled: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider
                link={newFetchMock({ ...externalService, warning: 'Invalid config we could not sync stuff' })}
            >
                <ExternalServiceEditPage
                    {...webProps}
                    routingPrefix="/site-admin"
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    externalServiceID="service123"
                    autoFocusForm={false}
                    externalServicesFromFile={true}
                    allowEditExternalServicesWithFile={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

EditingDisabled.storyName = 'External service config EXTSVC_CONFIG_FIlE set'
