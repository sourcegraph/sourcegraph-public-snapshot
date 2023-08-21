import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { ExecutionStatsBar } from './ExecutionStatsBar'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
    argTypes: {
        errored: {
            control: { type: 'number' },
            defaultValue: 0,
        },
        completed: {
            control: { type: 'number' },
            defaultValue: 7,
        },
        processing: {
            control: { type: 'number' },
            defaultValue: 4,
        },

        queued: {
            control: { type: 'number' },
            defaultValue: 14,
        },
        ignored: {
            control: { type: 'number' },
            defaultValue: 0,
        },
    },
}

export default config

export const ExecutionStatsBarStory: Story = args => (
    <WebStory>
        {props => (
            <ExecutionStatsBar
                {...props}
                errored={args.errored}
                completed={args.completed}
                processing={args.processing}
                queued={args.queued}
                ignored={args.ignored}
            />
        )}
    </WebStory>
)

ExecutionStatsBarStory.storyName = 'ExecutionStatsBar'
