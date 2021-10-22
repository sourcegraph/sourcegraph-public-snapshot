import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

import { WebhookAlert } from './WebhookAlert'

const { add } = storiesOf('web/batches/details/WebhookAlert', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const id = new Date().toString()

add('Site admin', () => (
    <WebStory>
        {() => (
            <WebhookAlert
                batchChange={{
                    id,
                    hasExternalServicesWithoutWebhooks: true,
                    externalServicesWithoutWebhooks: {
                        __typename: 'ExternalServiceConnection',
                        nodes: [
                            {
                                __typename: 'ExternalService',
                                id: 'ABCD',
                                kind: ExternalServiceKind.GITHUB,
                                displayName: 'GitHub',
                            },
                            {
                                __typename: 'ExternalService',
                                id: 'EFGH',
                                kind: ExternalServiceKind.GITLAB,
                                displayName: 'GitLab',
                            },
                        ],
                        pageInfo: {
                            hasNextPage: boolean('Has more external services', false),
                        },
                    },
                }}
            />
        )}
    </WebStory>
))

add('Regular user', () => (
    <WebStory>
        {() => (
            <WebhookAlert
                batchChange={{
                    id,
                    hasExternalServicesWithoutWebhooks: true,
                    externalServicesWithoutWebhooks: null,
                }}
            />
        )}
    </WebStory>
))

add('All external services have webhooks', () => (
    <WebStory>
        {() => (
            <WebhookAlert
                batchChange={{
                    id,
                    hasExternalServicesWithoutWebhooks: false,
                    externalServicesWithoutWebhooks: null,
                }}
            />
        )}
    </WebStory>
))
