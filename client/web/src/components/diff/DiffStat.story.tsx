import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../WebStory'

import { DiffStat, DiffStatSquares, DiffStatStack } from './DiffStat'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/diffs/DiffStat',
    decorators: [decorator],
    argTypes: {
        added: {
            type: 'number',
        },
        deleted: {
            type: 'number',
        },
    },
    args: {
        added: 10,
        deleted: 8,
    },
}

export default config

export const CollapsedCounts: StoryFn<React.ComponentProps<typeof DiffStat>> = args => (
    <WebStory>{() => <DiffStat {...args} />}</WebStory>
)

CollapsedCounts.storyName = 'Collapsed counts'

export const ExpandedCounts: StoryFn<React.ComponentProps<typeof DiffStat>> = args => (
    <WebStory>{() => <DiffStat {...args} expandedCounts={true} />}</WebStory>
)

ExpandedCounts.storyName = 'Expanded counts'

export const DiffStatSquaresStory: StoryFn<React.ComponentProps<typeof DiffStatSquares>> = args => (
    <WebStory>{() => <DiffStatSquares {...args} />}</WebStory>
)

DiffStatSquaresStory.storyName = 'DiffStatSquares'

export const DiffStatStackStory: StoryFn<React.ComponentProps<typeof DiffStatStack>> = args => (
    <WebStory>{() => <DiffStatStack {...args} />}</WebStory>
)

DiffStatStackStory.storyName = 'DiffStatStack'
