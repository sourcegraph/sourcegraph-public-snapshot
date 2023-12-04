import { action } from '@storybook/addon-actions'
import type { StoryFn, Meta, Decorator } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { PublishChangesetsModal } from './PublishChangesetsModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/PublishChangesetsModal',
    decorators: [decorator],
}

export default config

const publishChangesets = () => {
    action('PublishChangesets')
    return Promise.resolve()
}

export const Confirmation: StoryFn = () => (
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
)
