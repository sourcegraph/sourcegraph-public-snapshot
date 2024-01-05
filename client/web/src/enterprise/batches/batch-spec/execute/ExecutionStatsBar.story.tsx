import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { ExecutionStatsBar } from './ExecutionStatsBar'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute',
    decorators: [decorator],
    argTypes: {
        errored: {
            control: { type: 'number' },
        },
        completed: {
            control: { type: 'number' },
        },
        processing: {
            control: { type: 'number' },
        },

        queued: {
            control: { type: 'number' },
        },
        ignored: {
            control: { type: 'number' },
        },
    },
    args: {
        errored: 0,
        completed: 7,
        processing: 4,
        queued: 14,
        ignored: 0,
    },
}

export default config

export const ExecutionStatsBarStory: StoryFn = args => (
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
