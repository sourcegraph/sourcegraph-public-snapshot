import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'
import { ExternalServiceKind, WebhookFields } from '../graphql-operations'

import { WebhookInformation } from './WebhookInformation'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/WebhookInformation',
    decorators: [decorator],
}

export default config

export const WebhookDescription: Story = () => (
    <WebStory>{() => <WebhookInformation webhook={createWebhook()} />}</WebStory>
)

function createWebhook(): WebhookFields {
    return {
        __typename: 'Webhook',
        createdAt: '',
        id: '1',
        secret: 'secret-secret',
        updatedAt: '',
        url: 'sg.com/.api/webhooks/1aa2b42c-a14c-4aaa-b756-70c82e94d3e7',
        uuid: '1aa2b42c-a14c-4aaa-b756-70c82e94d3e7',
        codeHostKind: ExternalServiceKind.GITHUB,
        codeHostURN: 'github.com/repo1',
        createdBy: {
            username: 'alice',
            url: 'users/alice',
        },
        updatedBy: {
            username: 'alice',
            url: 'users/alice',
        },
    }
}

WebhookDescription.storyName = 'Webhook Information'
