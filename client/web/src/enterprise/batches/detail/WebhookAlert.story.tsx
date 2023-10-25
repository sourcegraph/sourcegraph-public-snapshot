import type { Meta, StoryFn, Decorator } from '@storybook/react'

import type { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'
import { BatchSpecSource } from '../../../graphql-operations'

import { WebhookAlert } from './WebhookAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/WebhookAlert',
    decorators: [decorator],
}

export default config

const id = new Date().toString()

const currentSpec = {
    id: 'specID1',
    originalInput: '',
    supersedingBatchSpec: null,
    source: BatchSpecSource.REMOTE,
    viewerBatchChangesCodeHosts: {
        totalCount: 0,
        nodes: [],
    },
    files: null,
    description: {
        __typename: 'BatchChangeDescription' as const,
        name: 'spec with ID 1',
    },
}

const batchChange = (totalCount: number, hasNextPage: boolean) => ({
    id,
    currentSpec: {
        ...currentSpec,
        codeHostsWithoutWebhooks: {
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
        },
    },
})

export const SiteAdmin: StoryFn = () => (
    <WebStory>{() => <WebhookAlert batchChange={batchChange(3, false)} isSiteAdmin={true} />}</WebStory>
)

SiteAdmin.storyName = 'Site admin'

export const RegularUser: StoryFn = () => (
    <WebStory>{() => <WebhookAlert batchChange={batchChange(3, false)} />}</WebStory>
)

RegularUser.storyName = 'Regular user'

export const RegularUserWithMoreThanThreeCodeHosts: StoryFn = () => (
    <WebStory>{() => <WebhookAlert batchChange={batchChange(4, true)} />}</WebStory>
)

RegularUserWithMoreThanThreeCodeHosts.storyName = 'Regular user with more than three code hosts'

export const AllCodeHostsHaveWebhooks: StoryFn = () => (
    <WebStory>
        {() => (
            <WebhookAlert
                batchChange={{
                    id,
                    currentSpec: {
                        ...currentSpec,
                        codeHostsWithoutWebhooks: {
                            nodes: [],
                            pageInfo: { hasNextPage: false },
                            totalCount: 0,
                        },
                    },
                }}
            />
        )}
    </WebStory>
)

AllCodeHostsHaveWebhooks.storyName = 'All code hosts have webhooks'
