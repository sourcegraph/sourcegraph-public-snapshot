import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { ClosedNotice } from './ClosedNotice'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/ClosedNotice',
    decorators: [decorator],
}

export default config

export const BatchChangeClosed: StoryFn = () => <WebStory>{() => <ClosedNotice closedAt="2021-02-02" />}</WebStory>

BatchChangeClosed.storyName = 'Batch change closed'
