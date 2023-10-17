import { action } from '@storybook/addon-actions'
import type { StoryFn, Meta, Decorator } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { MergeChangesetsModal } from './MergeChangesetsModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/MergeChangesetsModal',
    decorators: [decorator],
}

export default config

const mergeChangesets = () => {
    action('MergeChangesets')
    return Promise.resolve()
}

export const Confirmation: StoryFn = () => (
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
)
