import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subMinutes } from 'date-fns'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ExternalServiceKind, ExternalServiceSyncJobState } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { queryExternalServices as _queryExternalServices } from './backend'
import { ExternalServicesPage } from './ExternalServicesPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/External services/ExternalServicesPage',
    decorators: [decorator],
}

export default config

const queryExternalServices: typeof _queryExternalServices = () =>
    of({
        totalCount: 2,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [
            {
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
                    totalCount: 1,
                    pageInfo: { endCursor: null, hasNextPage: false },
                    nodes: [
                        {
                            __typename: 'ExternalServiceSyncJob',
                            failureMessage: null,
                            startedAt: subMinutes(new Date(), 25).toISOString(),
                            finishedAt: null,
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
                },
            },
        ],
    })

export const ListOfExternalServices: Story = () => (
    <WebStory>
        {webProps => (
            <ExternalServicesPage
                {...webProps}
                routingPrefix="/site-admin"
                afterDeleteRoute="/site-admin/after"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={{ id: '123' }}
                queryExternalServices={queryExternalServices}
                externalServicesFromFile={false}
                allowEditExternalServicesWithFile={false}
            />
        )}
    </WebStory>
)

ListOfExternalServices.storyName = 'List of external services'
