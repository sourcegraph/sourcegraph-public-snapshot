import { storiesOf } from '@storybook/react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

import { WebhookAlert } from './WebhookAlert'

const { add } = storiesOf('web/batches/details/WebhookAlert', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const id = new Date().toString()

const currentSpec = {
    id: 'specID1',
    originalInput: '',
    supersedingBatchSpec: null,
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
                    externalServiceURL: 'https://bitbucket.com/',
                },
            ],
            pageInfo: { hasNextPage },
            totalCount,
        },
    },
})

add('Site admin', () => (
    <WebStory>{() => <WebhookAlert batchChange={batchChange(3, false)} isSiteAdmin={true} />}</WebStory>
))

add('Regular user', () => <WebStory>{() => <WebhookAlert batchChange={batchChange(3, false)} />}</WebStory>)

add('Regular user with more than three code hosts', () => (
    <WebStory>{() => <WebhookAlert batchChange={batchChange(4, true)} />}</WebStory>
))

add('All code hosts have webhooks', () => (
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
))
