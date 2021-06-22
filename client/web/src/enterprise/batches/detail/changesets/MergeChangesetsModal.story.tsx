import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { MergeChangesetsModal } from './MergeChangesetsModal'

const { add } = storiesOf('web/batches/details/MergeChangesetsModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const changesetIDsFunction = () => Promise.resolve(['test-123', 'test-234'])
const mergeChangesets = () => {
    action('MergeChangesets')
    return Promise.resolve()
}

add('Confirmation', () => (
    <EnterpriseWebStory>
        {props => (
            <MergeChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={changesetIDsFunction}
                onCancel={noop}
                mergeChangesets={mergeChangesets}
            />
        )}
    </EnterpriseWebStory>
))
