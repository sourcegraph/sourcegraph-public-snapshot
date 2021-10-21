import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { MergeChangesetsModal } from './MergeChangesetsModal'

const { add } = storiesOf('web/batches/details/MergeChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const mergeChangesets = () => {
    action('MergeChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <WebStory>
        {props => (
            <MergeChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                mergeChangesets={mergeChangesets}
            />
        )}
    </WebStory>
))
