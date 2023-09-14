import { action } from '@storybook/addon-actions'
import type { Story, Meta, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { MergeChangesetsModal } from './MergeChangesetsModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/MergeChangesetsModal',
    decorators: [decorator],
}

export default config

const mergeChangesets = () => {
    action('MergeChangesets')
    return Promise.resolve()
}

export const Confirmation: Story = () => (
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
