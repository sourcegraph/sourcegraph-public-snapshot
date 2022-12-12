import { formatRFC3339 } from 'date-fns'
import { of } from 'rxjs'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { queryExternalServices as _queryExternalServices } from '../components/externalServices/backend'
import { ListExternalServiceFields, WebhookFields } from '../graphql-operations'

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
        webhookURL: null,
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

export const queryExternalServices: typeof _queryExternalServices = () =>
    of({
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
            createExternalService(ExternalServiceKind.GERRIT, 'https://gerrit.com'),
            createExternalService(ExternalServiceKind.GITLAB, 'https://gitlab.com'),
            createExternalService(ExternalServiceKind.GITOLITE, 'https://gitolite.com'),
            createExternalService(ExternalServiceKind.GOMODULES, 'https://gomodules.com'),
            createExternalService(ExternalServiceKind.JVMPACKAGES, 'https://jvmpackages.com'),
            createExternalService(ExternalServiceKind.NPMPACKAGES, 'https://npmpackages.com'),
            createExternalService(ExternalServiceKind.OTHER, 'https://other.com'),
            createExternalService(ExternalServiceKind.PAGURE, 'https://pagure.com'),
            createExternalService(ExternalServiceKind.PERFORCE, 'https://perforce.com'),
            createExternalService(ExternalServiceKind.PHABRICATOR, 'https://phabricator.com'),
            createExternalService(ExternalServiceKind.PYTHONPACKAGES, 'https://pythonpackages.com'),
            createExternalService(ExternalServiceKind.RUSTPACKAGES, 'https://rustpackages.com'),
            createExternalService(ExternalServiceKind.RUBYPACKAGES, 'https://rubypackages.com'),
        ],
    })
