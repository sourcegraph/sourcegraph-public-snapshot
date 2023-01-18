import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subMinutes } from 'date-fns'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { ExternalServiceKind, ExternalServiceSyncJobState } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { EXTERNAL_SERVICES } from './backend'
import { ExternalServicesPage } from './ExternalServicesPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/External services/ExternalServicesPage',
    decorators: [decorator],
}

export default config

export const ListOfExternalServices: Story = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(EXTERNAL_SERVICES),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                            result: {
                                data: {
                                    externalServices: {
                                        totalCount: 2,
                                        pageInfo: {
                                            endCursor: null,
                                            hasNextPage: false,
                                        },
                                        nodes: [
                                            {
                                                __typename: 'ExternalService',
                                                id: 'service1',
                                                kind: ExternalServiceKind.GITHUB,
                                                displayName: 'GitHub #1',
                                                config: '{"githubconfig":true}',
                                                warning: null,
                                                lastSyncError: null,
                                                repoCount: 0,
                                                lastSyncAt: null,
                                                nextSyncAt: null,
                                                updatedAt: '2021-03-15T19:39:11Z',
                                                createdAt: '2021-03-15T19:39:11Z',
                                                namespace: null,
                                                webhookURL: null,
                                                hasConnectionCheck: true,
                                                syncJobs: {
                                                    __typename: 'ExternalServiceSyncJobConnection',
                                                    totalCount: 1,
                                                    pageInfo: { endCursor: null, hasNextPage: false },
                                                    nodes: [
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
                                                    ],
                                                },
                                            },
                                            {
                                                __typename: 'ExternalService',
                                                id: 'service2',
                                                kind: ExternalServiceKind.GITHUB,
                                                displayName: 'GitHub #2',
                                                config: '{"githubconfig":true}',
                                                warning: null,
                                                lastSyncError: null,
                                                repoCount: 0,
                                                lastSyncAt: null,
                                                nextSyncAt: null,
                                                updatedAt: '2021-03-15T19:39:11Z',
                                                createdAt: '2021-03-15T19:39:11Z',
                                                namespace: {
                                                    id: 'someuser-id',
                                                    namespaceName: 'johndoe',
                                                    url: '/users/johndoe',
                                                },
                                                webhookURL: null,
                                                hasConnectionCheck: false,
                                                syncJobs: {
                                                    __typename: 'ExternalServiceSyncJobConnection',
                                                    totalCount: 1,
                                                    pageInfo: { endCursor: null, hasNextPage: false },
                                                    nodes: [
                                                        {
                                                            __typename: 'ExternalServiceSyncJob',
                                                            failureMessage: null,
                                                            startedAt: subMinutes(new Date(), 25).toISOString(),
                                                            finishedAt: null,
                                                            id: 'SYNCJOB2',
                                                            state: ExternalServiceSyncJobState.COMPLETED,
                                                            reposSynced: 5,
                                                            repoSyncErrors: 0,
                                                            reposAdded: 5,
                                                            reposDeleted: 0,
                                                            reposModified: 0,
                                                            reposUnmodified: 0,
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                            },
                        },
                    ])
                }
            >
                <ExternalServicesPage
                    {...webProps}
                    routingPrefix="/site-admin"
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    authenticatedUser={{ id: '123' }}
                    externalServicesFromFile={false}
                    allowEditExternalServicesWithFile={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfExternalServices.storyName = 'List of external services'
