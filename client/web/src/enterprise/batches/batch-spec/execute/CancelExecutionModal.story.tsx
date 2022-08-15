import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { Text } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'

import { CancelExecutionModal } from './CancelExecutionModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
}

export default config

export const CancelExecutionModalStory: Story = () => (
    <WebStory>
        {props => (
            <CancelExecutionModal
                {...props}
                modalBody={<Text>Are you sure you want to cancel the current execution?</Text>}
                isOpen={true}
                isLoading={boolean('isLoading', false)}
                onCancel={noop}
                onConfirm={noop}
            />
        )}
    </WebStory>
)

CancelExecutionModalStory.storyName = 'CancelExecutionModal'
