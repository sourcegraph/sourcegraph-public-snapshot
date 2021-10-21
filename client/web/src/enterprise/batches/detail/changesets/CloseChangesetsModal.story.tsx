import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { CloseChangesetsModal } from './CloseChangesetsModal'

const { add } = storiesOf('web/batches/details/CloseChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const closeChangesets = () => {
    action('CloseChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <WebStory>
        {props => (
            <CloseChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                closeChangesets={closeChangesets}
            />
        )}
    </WebStory>
))
