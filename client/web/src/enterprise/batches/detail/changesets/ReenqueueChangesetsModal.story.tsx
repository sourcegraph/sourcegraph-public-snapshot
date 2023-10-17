import { action } from '@storybook/addon-actions'
import type { StoryFn, Meta, Decorator } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { ReenqueueChangesetsModal } from './ReenqueueChangesetsModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/ReenqueueChangesetsModal',
    decorators: [decorator],
}

export default config

const reenqueueChangesets = () => {
    action('ReenqueueChangesets')
    return Promise.resolve()
}

export const Confirmation: StoryFn = () => (
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
)
