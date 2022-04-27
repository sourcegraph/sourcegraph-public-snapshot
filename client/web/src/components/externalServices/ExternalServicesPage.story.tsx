import { storiesOf } from '@storybook/react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ExternalServiceKind } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { queryExternalServices as _queryExternalServices } from './backend'
import { ExternalServicesPage } from './ExternalServicesPage'

const { add } = storiesOf('web/External services/ExternalServicesPage', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

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
                displayName: 'GitHub.com',
                config: '{"githubconfig":true}',
                warning: null,
                lastSyncError: null,
                repoCount: 0,
                lastSyncAt: null,
                nextSyncAt: null,
                updatedAt: '2021-03-15T19:39:11Z',
                createdAt: '2021-03-15T19:39:11Z',
                namespace: null,
                grantedScopes: [],
            },
            {
                id: 'service2',
                kind: ExternalServiceKind.GITHUB,
                displayName: 'GitHub.com',
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
                grantedScopes: [],
            },
        ],
    })

add('List of external services', () => (
    <WebStory>
        {webProps => (
            <ExternalServicesPage
                {...webProps}
                routingPrefix="/site-admin"
                afterDeleteRoute="/site-admin/after"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={{ id: '123' }}
                queryExternalServices={queryExternalServices}
            />
        )}
    </WebStory>
))
