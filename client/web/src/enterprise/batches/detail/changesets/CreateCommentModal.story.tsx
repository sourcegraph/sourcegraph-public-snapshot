import { action } from '@storybook/addon-actions'
import type { Meta, Story, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { CreateCommentModal } from './CreateCommentModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/CreateCommentModal',
    decorators: [decorator],
}

export default config

const createChangesetCommentsAction = () => {
    action('CreateChangesetComments')
    return Promise.resolve()
}

export const Confirmation: Story = () => (
    <WebStory>
        {props => (
            <CreateCommentModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                createChangesetComments={createChangesetCommentsAction}
            />
        )}
    </WebStory>
)
