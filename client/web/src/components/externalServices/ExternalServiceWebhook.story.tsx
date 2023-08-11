import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { ExternalServiceKind } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { ExternalServiceWebhook } from './ExternalServiceWebhook'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="p-3 container">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/External services/ExternalServiceWebhook',
    decorators: [decorator],
}

export default config

export const BitbucketServer: Story = () => (
    <ExternalServiceWebhook
        externalService={{ webhookURL: 'http://test.test/webhook', kind: ExternalServiceKind.BITBUCKETSERVER }}
    />
)

export const GitHub: Story = () => (
    <ExternalServiceWebhook
        externalService={{ webhookURL: 'http://test.test/webhook', kind: ExternalServiceKind.GITHUB }}
    />
)

export const GitLab: Story = () => (
    <ExternalServiceWebhook
        externalService={{ webhookURL: 'http://test.test/webhook', kind: ExternalServiceKind.GITLAB }}
    />
)

GitLab.parameters = {
    chromatic: {
        // Visually the same as GitHub
        disable: true,
    },
}
