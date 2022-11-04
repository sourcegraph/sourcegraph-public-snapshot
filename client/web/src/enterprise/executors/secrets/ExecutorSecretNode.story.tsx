import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays, subHours } from 'date-fns'

import { WebStory } from '../../../components/WebStory'
import { ExecutorSecretFields, ExecutorSecretScope } from '../../../graphql-operations'

import { ExecutorSecretNode } from './ExecutorSecretNode'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/ExecutorSecretNode',
    decorators: [decorator],
}

export default config

const secret: ExecutorSecretFields = {
    __typename: 'ExecutorSecret',
    id: 'secret1',
    creator: {
        __typename: 'User',
        username: 'test',
        displayName: 'Test user',
        id: 'testID',
        url: '/users/test',
    },
    key: 'SG_TOKEN',
    scope: ExecutorSecretScope.BATCHES,
    // Global secret.
    namespace: null,
    createdAt: subDays(new Date(), 1).toISOString(),
    updatedAt: subHours(new Date(), 12).toISOString(),
}

export const Overview: Story = () => (
    <WebStory>{props => <ExecutorSecretNode {...props} node={secret} refetchAll={() => {}} />}</WebStory>
)
