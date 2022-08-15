import { DecoratorFn, Story, Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesChangelogAlert } from './BatchChangesChangelogAlert'

const decorator: DecoratorFn = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangesChangelogAlert',
    decorators: [decorator],
}

export default config

export const Changelog: Story = () => <WebStory>{() => <BatchChangesChangelogAlert />}</WebStory>
