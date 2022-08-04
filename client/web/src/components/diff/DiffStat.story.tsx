import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../WebStory'

import { DiffStat, DiffStatSquares, DiffStatStack } from './DiffStat'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/diffs/DiffStat',
    decorators: [decorator],
    argTypes: {
        added: {
            type: 'number',
            defaultValue: 10,
        },
        changed: {
            type: 'number',
            defaultValue: 4,
        },
        deleted: {
            type: 'number',
            defaultValue: 8,
        },
    },
}

export default config

export const CollapsedCounts: Story<React.ComponentProps<typeof DiffStat>> = args => (
    <WebStory>{() => <DiffStat {...args} />}</WebStory>
)

CollapsedCounts.storyName = 'Collapsed counts'

export const ExpandedCounts: Story = args => (
    <WebStory>
        {() => <DiffStat added={args.added} changed={args.changed} deleted={args.deleted} expandedCounts={true} />}
    </WebStory>
)

ExpandedCounts.storyName = 'Expanded counts'

export const DiffStatSquaresStory: Story = args => (
    <WebStory>{() => <DiffStatSquares added={args.added} changed={args.changed} deleted={args.deleted} />}</WebStory>
)

DiffStatSquaresStory.storyName = 'DiffStatSquares'

export const DiffStatStackStory: Story = args => (
    <WebStory>{() => <DiffStatStack added={args.added} changed={args.changed} deleted={args.deleted} />}</WebStory>
)

DiffStatStackStory.storyName = 'DiffStatStack'
