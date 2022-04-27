import { storiesOf } from '@storybook/react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ExternalServiceKind } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { fetchExternalService as _fetchExternalService } from './backend'
import { ExternalServicePage } from './ExternalServicePage'

const { add } = storiesOf('web/External services/ExternalServicePage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so Monaco has some time to get syntax highlighting prepared.
            delay: 2000,
        },
    })

const externalService = {
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

const fetchExternalService: typeof _fetchExternalService = () => of(externalService)

const fetchExternalServiceWithInvalidConfigURL: typeof _fetchExternalService = () =>
    of({ ...externalService, config: '{"url": "invalid-url"}' })

add('View external service config', () => (
    <WebStory>
        {webProps => (
            <ExternalServicePage
                {...webProps}
                afterUpdateRoute="/site-admin/after"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                externalServiceID="service123"
                fetchExternalService={fetchExternalService}
                autoFocusForm={false}
            />
        )}
    </WebStory>
))

add('External service config with invalid url', () => (
    <WebStory>
        {webProps => (
            <ExternalServicePage
                {...webProps}
                afterUpdateRoute="/site-admin/after"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                externalServiceID="service123"
                fetchExternalService={fetchExternalServiceWithInvalidConfigURL}
                autoFocusForm={false}
            />
        )}
    </WebStory>
))
