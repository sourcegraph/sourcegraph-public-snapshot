import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'
import { ExecutorSecretScope } from '../../../graphql-operations'

import { AddSecretModal } from './AddSecretModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/AddSecretModal',
    decorators: [decorator],
    parameters: {},
}

export default config

export const GitHub: StoryFn = () => (
    <WebStory>
        {props => (
            <AddSecretModal
                {...props}
                namespaceID="user-id-1"
                scope={ExecutorSecretScope.BATCHES}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
)

GitHub.storyName = 'Add secret'

export const DockerAuthConfig: StoryFn = () => (
    <WebStory>
        {props => (
            <AddSecretModal
                {...props}
                namespaceID="user-id-1"
                scope={ExecutorSecretScope.BATCHES}
                afterCreate={noop}
                onCancel={noop}
                initialKey="DOCKER_AUTH_CONFIG"
            />
        )}
    </WebStory>
)

DockerAuthConfig.storyName = 'Docker auth config'
