import { DecoratorFn, Story, Meta } from '@storybook/react'
import { subMinutes } from 'date-fns'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { ExternalServiceFields, ExternalServiceKind, ExternalServiceSyncJobState } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { FETCH_EXTERNAL_SERVICE, queryExternalServiceSyncJobs as _queryExternalServiceSyncJobs } from './backend'
import { ExternalServicePage } from './ExternalServicePage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/External services/ExternalServicePage',
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
    grantedScopes: [],
    namespace: {
        id: 'userid',
        namespaceName: 'johndoe',
        url: '/users/johndoe',
    },
}

const queryExternalServiceSyncJobs: typeof _queryExternalServiceSyncJobs = () =>
    of({
        totalCount: 4,
        pageInfo: { endCursor: null, hasNextPage: false },
        nodes: [
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: null,
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: null,
                id: 'SYNCJOB1',
                state: ExternalServiceSyncJobState.CANCELING,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: null,
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: null,
                id: 'SYNCJOB1',
                state: ExternalServiceSyncJobState.PROCESSING,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: 'Very bad error syncing with the code host.',
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: subMinutes(new Date(), 25).toISOString(),
                id: 'SYNCJOB1',
                state: ExternalServiceSyncJobState.FAILED,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: null,
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: subMinutes(new Date(), 25).toISOString(),
                id: 'SYNCJOB1',
                state: ExternalServiceSyncJobState.COMPLETED,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
        ],
    })

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
                <ExternalServicePage
                    {...webProps}
                    queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                    afterUpdateRoute="/site-admin/after"
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
                <ExternalServicePage
                    {...webProps}
                    queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                    afterUpdateRoute="/site-admin/after"
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
                <ExternalServicePage
                    {...webProps}
                    queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                    afterUpdateRoute="/site-admin/after"
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
                <ExternalServicePage
                    {...webProps}
                    queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                    afterUpdateRoute="/site-admin/after"
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
