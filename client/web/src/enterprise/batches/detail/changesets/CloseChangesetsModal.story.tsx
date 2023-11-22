import { action } from '@storybook/addon-actions'
import type { StoryFn, Decorator, Meta } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { CloseChangesetsModal } from './CloseChangesetsModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/CloseChangesetsModal',
    decorators: [decorator],
}

export default config

const closeChangesets = () => {
    action('CloseChangesets')
    return Promise.resolve()
}

export const Confirmation: StoryFn = () => (
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
)
