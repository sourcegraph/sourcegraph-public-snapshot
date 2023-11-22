import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { subDays, subHours } from 'date-fns'
import { noop } from 'lodash'

import { ExecutorSecretScope } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'
import type { ExecutorSecretFields } from '../../../graphql-operations'

import { RemoveSecretModal } from './RemoveSecretModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/RemoveSecretModal',
    decorators: [decorator],
    parameters: {
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    },
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

export const Confirm: StoryFn = () => (
    <WebStory>{props => <RemoveSecretModal {...props} secret={secret} afterDelete={noop} onCancel={noop} />}</WebStory>
)

Confirm.storyName = 'Confirm delete'
