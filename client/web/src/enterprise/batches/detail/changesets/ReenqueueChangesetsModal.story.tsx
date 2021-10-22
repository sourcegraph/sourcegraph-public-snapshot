import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { ReenqueueChangesetsModal } from './ReenqueueChangesetsModal'

const { add } = storiesOf('web/batches/details/ReenqueueChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const reenqueueChangesets = () => {
    action('ReenqueueChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <WebStory>
        {props => (
            <ReenqueueChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                reenqueueChangesets={reenqueueChangesets}
            />
        )}
    </WebStory>
))
