import { Meta, Story, DecoratorFn } from '@storybook/react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

import { WebhookAlert } from './WebhookAlert'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/WebhookAlert',
    decorators: [decorator],
}

export default config

const id = new Date().toString()

const codeHostsWithoutWebhooks = (totalCount: number, hasNextPage: boolean) => ({
    nodes: [
        {
            externalServiceKind: 'GITHUB' as ExternalServiceKind,
            externalServiceURL: 'https://github.com/',
        },
        {
            externalServiceKind: 'GITLAB' as ExternalServiceKind,
            externalServiceURL: 'https://gitlab.com/',
        },
        {
            externalServiceKind: 'BITBUCKETSERVER' as ExternalServiceKind,
            externalServiceURL: 'https://bitbucket.org/',
        },
    ],
    pageInfo: { hasNextPage },
    totalCount,
})

export const SiteAdmin: Story = () => (
    <WebStory>
        {() => (
            <WebhookAlert
                batchChangeID={id}
                codeHostsWithoutWebhooks={codeHostsWithoutWebhooks(3, false)}
                isSiteAdmin={true}
            />
        )}
    </WebStory>
)

SiteAdmin.storyName = 'Site admin'

export const RegularUser: Story = () => (
    <WebStory>
        {() => <WebhookAlert batchChangeID={id} codeHostsWithoutWebhooks={codeHostsWithoutWebhooks(3, false)} />}
    </WebStory>
)

RegularUser.storyName = 'Regular user'

export const RegularUserWithMoreThanThreeCodeHosts: Story = () => (
    <WebStory>
        {() => <WebhookAlert batchChangeID={id} codeHostsWithoutWebhooks={codeHostsWithoutWebhooks(4, true)} />}
    </WebStory>
)

RegularUserWithMoreThanThreeCodeHosts.storyName = 'Regular user with more than three code hosts'

export const AllCodeHostsHaveWebhooks: Story = () => (
    <WebStory>
        {() => (
            <WebhookAlert
                batchChangeID={id}
                codeHostsWithoutWebhooks={{
                    nodes: [],
                    pageInfo: { hasNextPage: false },
                    totalCount: 0,
                }}
            />
        )}
    </WebStory>
)

AllCodeHostsHaveWebhooks.storyName = 'All code hosts have webhooks'
