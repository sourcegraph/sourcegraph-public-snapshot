import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../WebStory'

import { DiffStat, DiffStatSquares, DiffStatStack } from './DiffStat'

const getSharedKnobs = () => ({
    added: number('Added', 10),
    changed: number('Changed', 4),
    deleted: number('Deleted', 8),
})

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/diffs/DiffStat',
    decorators: [decorator],
}

export default config

export const CollapsedCounts: Story = () => {
    const stats = getSharedKnobs()
    return <WebStory>{() => <DiffStat {...stats} />}</WebStory>
}

CollapsedCounts.storyName = 'Collapsed counts'

export const ExpandedCounts: Story = () => {
    const stats = getSharedKnobs()
    return <WebStory>{() => <DiffStat {...stats} expandedCounts={true} />}</WebStory>
}

ExpandedCounts.storyName = 'Expanded counts'

export const DiffStatSquaresStory: Story = () => {
    const stats = getSharedKnobs()
    return <WebStory>{() => <DiffStatSquares {...stats} />}</WebStory>
}

DiffStatSquaresStory.storyName = 'DiffStatSquares'

export const DiffStatStackStory: Story = () => {
    const stats = getSharedKnobs()
    return <WebStory>{() => <DiffStatStack {...stats} />}</WebStory>
}

DiffStatStackStory.storyName = 'DiffStatStack'
