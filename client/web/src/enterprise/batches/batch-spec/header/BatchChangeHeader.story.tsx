import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { BatchChangeHeader } from './BatchChangeHeader'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/header/BatchChangeHeader',
    decorators: [decorator],
}

export default config

export const CreateNewBatchChange: Story = () => (
    <WebStory>{props => <BatchChangeHeader {...props} title={{ text: 'Create batch change' }} />}</WebStory>
)

CreateNewBatchChange.storyName = 'creating a new batch change'

export const BatchChangeExists: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeHeader
                {...props}
                namespace={{ to: '/users/my-username', text: 'my-username' }}
                title={{ to: '/users/my-username/batch-changes/my-batch-change', text: 'my-batch-change' }}
                description="This is a description of my batch change."
            />
        )}
    </WebStory>
)

BatchChangeExists.storyName = 'batch change already exists'
