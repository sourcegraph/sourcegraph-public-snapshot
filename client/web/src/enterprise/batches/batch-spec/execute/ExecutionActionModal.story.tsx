import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { Text } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'

import { ExecutionActionModal } from './ExecutionActionModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
    argTypes: {
        isLoading: {
            control: {
                type: 'boolean',
            },
            defaultValue: false,
        },
    },
}

export default config

export const ExecutionActionModalStory: Story = args => (
    <WebStory>
        {props => (
            <ExecutionActionModal
                {...props}
                modalHeader="Cancel execution"
                confirmLabel="Cancel"
                modalBody={<Text>Are you sure you want to cancel the current execution?</Text>}
                isOpen={true}
                isLoading={args.isLoading}
                onCancel={noop}
                onConfirm={noop}
            />
        )}
    </WebStory>
)

ExecutionActionModalStory.storyName = 'ExecutionActionModal'
