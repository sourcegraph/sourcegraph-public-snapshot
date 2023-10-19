import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { Text } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'

import { CancelExecutionModal } from './CancelExecutionModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
    argTypes: {
        isLoading: {
            control: {
                type: 'boolean',
            },
        },
    },
    args: {
        isLoading: false,
    },
}

export default config

export const CancelExecutionModalStory: StoryFn = args => (
    <WebStory>
        {props => (
            <CancelExecutionModal
                {...props}
                modalBody={<Text>Are you sure you want to cancel the current execution?</Text>}
                isOpen={true}
                isLoading={args.isLoading}
                onCancel={noop}
                onConfirm={noop}
            />
        )}
    </WebStory>
)

CancelExecutionModalStory.storyName = 'CancelExecutionModal'
