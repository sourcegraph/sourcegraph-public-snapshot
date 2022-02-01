import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { CreateCommentModal } from './CreateCommentModal'

const { add } = storiesOf('web/batches/details/CreateCommentModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const createChangesetCommentsAction = () => {
    action('CreateChangesetComments')
    return Promise.resolve()
}

add('Confirmation', () => (
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
))
