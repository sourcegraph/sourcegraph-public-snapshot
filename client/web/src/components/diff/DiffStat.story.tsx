import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../WebStory'

import { DiffStat, DiffStatSquares, DiffStatStack } from './DiffStat'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

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

export const CollapsedCounts: Story<React.ComponentProps<typeof DiffStat>> = args => (
    <WebStory>{() => <DiffStat {...args} />}</WebStory>
)

CollapsedCounts.storyName = 'Collapsed counts'

export const ExpandedCounts: Story<React.ComponentProps<typeof DiffStat>> = args => (
    <WebStory>{() => <DiffStat {...args} expandedCounts={true} />}</WebStory>
)

ExpandedCounts.storyName = 'Expanded counts'

export const DiffStatSquaresStory: Story<React.ComponentProps<typeof DiffStatSquares>> = args => (
    <WebStory>{() => <DiffStatSquares {...args} />}</WebStory>
)

DiffStatSquaresStory.storyName = 'DiffStatSquares'

export const DiffStatStackStory: Story<React.ComponentProps<typeof DiffStatStack>> = args => (
    <WebStory>{() => <DiffStatStack {...args} />}</WebStory>
)

DiffStatStackStory.storyName = 'DiffStatStack'
