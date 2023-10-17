import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { subDays, subHours } from 'date-fns'

import { WebStory } from '../../../components/WebStory'
import { type ExecutorSecretFields, ExecutorSecretScope } from '../../../graphql-operations'

import { ExecutorSecretNode } from './ExecutorSecretNode'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

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
    overwritesGlobalSecret: false,
    createdAt: subDays(new Date(), 1).toISOString(),
    updatedAt: subHours(new Date(), 12).toISOString(),
}

export const Overview: StoryFn = () => (
    <WebStory>
        {props => <ExecutorSecretNode {...props} namespaceID={null} node={secret} refetchAll={() => {}} />}
    </WebStory>
)

const overwrittenSecret: ExecutorSecretFields = {
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
    namespace: {
        __typename: 'User',
        id: 'namespace',
        namespaceName: 'user',
        url: '/users/user',
    },
    overwritesGlobalSecret: true,
    createdAt: subDays(new Date(), 1).toISOString(),
    updatedAt: subHours(new Date(), 12).toISOString(),
}

export const OverwritesGlobal: StoryFn = () => (
    <WebStory>
        {props => <ExecutorSecretNode {...props} namespaceID={null} node={overwrittenSecret} refetchAll={() => {}} />}
    </WebStory>
)

OverwritesGlobal.storyName = 'Overwrites global secret'
