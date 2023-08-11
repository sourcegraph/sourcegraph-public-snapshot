import type { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { ClosedNotice } from './ClosedNotice'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/ClosedNotice',
    decorators: [decorator],
}

export default config

export const BatchChangeClosed: Story = () => <WebStory>{() => <ClosedNotice closedAt="2021-02-02" />}</WebStory>

BatchChangeClosed.storyName = 'Batch change closed'
