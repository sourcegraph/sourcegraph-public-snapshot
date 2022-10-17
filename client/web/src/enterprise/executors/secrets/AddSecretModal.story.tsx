import { DecoratorFn, Story, Meta } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'
import { ExecutorSecretScope } from '../../../graphql-operations'

import { AddSecretModal } from './AddSecretModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/AddSecretModal',
    decorators: [decorator],
    parameters: {
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    },
}

export default config

export const GitHub: Story = () => (
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
