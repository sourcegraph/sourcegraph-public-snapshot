import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'
import { of, throwError } from 'rxjs'

import { asError } from '@sourcegraph/common'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { queryExternalServices as _queryExternalServices } from '../components/externalServices/backend'
import { WebStory } from '../components/WebStory'
import { ListExternalServiceFields } from '../graphql-operations'

import { SiteAdminWebhookCreatePage } from './SiteAdminWebhookCreatePage'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/SiteAdminWebhookCreatePage',
    decorators: [decorator],
}

export default config

const queryExternalServices: typeof _queryExternalServices = () =>
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

function createExternalService(kind: ExternalServiceKind, url: string): ListExternalServiceFields {
    return {
        id: 'service1',
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

export const WebhookCreatePage: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <SiteAdminWebhookCreatePage
                    match={{} as any}
                    history={H.createMemoryHistory()}
                    location={{} as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    queryExternalServices={queryExternalServices}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

WebhookCreatePage.storyName = 'Create webhook'

const queryExternalServicesError: typeof _queryExternalServices = () => throwError(asError('oops'))

export const WebhookCreatePageWithError: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <SiteAdminWebhookCreatePage
                    match={{} as any}
                    history={H.createMemoryHistory()}
                    location={{} as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    queryExternalServices={queryExternalServicesError}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

WebhookCreatePageWithError.storyName = 'Error during external services fetch'
