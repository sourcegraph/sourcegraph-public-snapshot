import { select, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../components/WebStory'

import { BatchSpecStateBadge } from './BatchSpecStateBadge'

const decorator: DecoratorFn = story => <div className="p-3">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/BatchSpecStateBadge',
    decorators: [decorator, withKnobs],
}

export default config

export const BatchSpecStateBadgeStory: Story = () => (
    <WebStory>
        {props => (
            <BatchSpecStateBadge
                state={select(
                    'State',
                    [
                        BatchSpecState.PENDING,
                        BatchSpecState.QUEUED,
                        BatchSpecState.PROCESSING,
                        BatchSpecState.CANCELED,
                        BatchSpecState.CANCELING,
                        BatchSpecState.FAILED,
                        BatchSpecState.COMPLETED,
                    ],
                    BatchSpecState.PENDING
                )}
                {...props}
            />
        )}
    </WebStory>
)

BatchSpecStateBadgeStory.storyName = 'BatchSpecStateBadge'
