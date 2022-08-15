import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { ExecutionStatsBar } from './ExecutionStatsBar'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
}

export default config

export const ExecutionStatsBarStory: Story = () => (
    <WebStory>
        {props => (
            <ExecutionStatsBar
                {...props}
                errored={number('errored', 0)}
                ignored={number('ignored', 0)}
                queued={number('queued', 14)}
                processing={number('processing', 4)}
                completed={number('completed', 7)}
            />
        )}
    </WebStory>
)

ExecutionStatsBarStory.storyName = 'ExecutionStatsBar'
