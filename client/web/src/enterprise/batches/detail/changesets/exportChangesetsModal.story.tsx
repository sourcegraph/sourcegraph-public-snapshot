import { action } from '@storybook/addon-actions'
import { Story, Meta, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../../components/WebStory'

import { ExportChangesetsModal } from './ExportChangesetsModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/ExportChangesetsModal',
    decorators: [decorator],
}

export default config

const exportChangesets = () => {
    action('ExportChangesets')
    return Promise.resolve()
}

export const Confirmation: Story = () => (
    <WebStory>
        {props => (
            <ExportChangesetsModal
                {...props}
                afterCreate={noop}
                batchChangeID="test-123"
                changesetIDs={['test-123', 'test-234']}
                onCancel={noop}
                exportChangesets={exportChangesets}
            />
        )}
    </WebStory>
)
