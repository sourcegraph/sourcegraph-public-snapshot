import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { PublishChangesetsModal } from './PublishChangesetsModal'

const { add } = storiesOf('web/batches/details/PublishChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const publishChangesets = () => {
    action('PublishChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <WebStory>
        {props => (
            <PublishChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                publishChangesets={publishChangesets}
            />
        )}
    </WebStory>
))
