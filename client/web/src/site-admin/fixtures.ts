import { formatRFC3339, subMinutes } from 'date-fns'

import { type ExternalServiceKind, ExternalServiceSyncJobState } from '@sourcegraph/shared/src/graphql-operations'

import type { ListExternalServiceFields, WebhookFields } from '../graphql-operations'

export const TIMESTAMP_MOCK = new Date(2021, 10, 8, 16, 40, 30)

export function createExternalService(kind: ExternalServiceKind, url: string): ListExternalServiceFields {
    return {
        __typename: 'ExternalService',
        id: `service-${url}`,
        kind,
        displayName: `${kind}-123`,
        config: `{"url": "${url}"}`,
        warning: null,
        lastSyncError: null,
        repoCount: 0,
        lastSyncAt: null,
        nextSyncAt: null,
        updatedAt: '2021-03-15T19:39:11Z',
        createdAt: '2021-03-15T19:39:11Z',
        creator: {
            username: 'alice',
            url: '/users/alice',
        },
        lastUpdater: {
            username: 'alice',
            url: '/users/alice',
        },
        rateLimiterState: {
            __typename: 'RateLimiterState',
            currentCapacity: 10,
            burst: 10,
            limit: 5000,
            interval: 1,
            lastReplenishment: '2021-03-15T19:39:11Z',
            infinite: false,
        },
        webhookURL: null,
        hasConnectionCheck: true,
        unrestricted: false,
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
    }
}

export function createWebhookMock(kind: ExternalServiceKind, urn: string): WebhookFields {
    return {
        __typename: 'Webhook',
        createdAt: formatRFC3339(TIMESTAMP_MOCK),
        id: '1',
        name: 'sgprod.bitbucket.org commit push webhook',
        secret: 'secret-secret',
        updatedAt: formatRFC3339(TIMESTAMP_MOCK),
        url: 'sg.com/.api/webhooks/1aa2b42c-a14c-4aaa-b756-70c82e94d3e7',
        uuid: '1aa2b42c-a14c-4aaa-b756-70c82e94d3e7',
        codeHostKind: kind,
        codeHostURN: urn,
        updatedBy: {
            username: 'alice',
            url: '/users/alice',
        },
        createdBy: {
            username: 'alice',
            url: '/users/alice',
        },
    }
}
