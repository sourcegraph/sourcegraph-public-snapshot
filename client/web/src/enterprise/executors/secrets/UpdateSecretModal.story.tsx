import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { subDays, subHours } from 'date-fns'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'
import { ExecutorSecretScope } from '../../../graphql-operations'

import { UpdateSecretModal } from './UpdateSecretModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/UpdateSecretModal',
    decorators: [decorator],
    parameters: {
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    },
}

export default config

export const Update: StoryFn = () => (
    <WebStory>
        {props => (
            <UpdateSecretModal
                {...props}
                secret={{
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
                    overwritesGlobalSecret: false,
                    // Global secret.
                    namespace: null,
                    createdAt: subDays(new Date(), 1).toISOString(),
                    updatedAt: subHours(new Date(), 12).toISOString(),
                }}
                onCancel={noop}
                afterUpdate={noop}
            />
        )}
    </WebStory>
)

export const DockerAuthConfig: StoryFn = () => (
    <WebStory>
        {props => (
            <UpdateSecretModal
                {...props}
                secret={{
                    __typename: 'ExecutorSecret',
                    id: 'secret1',
                    creator: {
                        __typename: 'User',
                        username: 'test',
                        displayName: 'Test user',
                        id: 'testID',
                        url: '/users/test',
                    },
                    key: 'DOCKER_AUTH_CONFIG',
                    scope: ExecutorSecretScope.BATCHES,
                    overwritesGlobalSecret: false,
                    // Global secret.
                    namespace: null,
                    createdAt: subDays(new Date(), 1).toISOString(),
                    updatedAt: subHours(new Date(), 12).toISOString(),
                }}
                onCancel={noop}
                afterUpdate={noop}
            />
        )}
    </WebStory>
)

DockerAuthConfig.storyName = 'Docker auth config'
