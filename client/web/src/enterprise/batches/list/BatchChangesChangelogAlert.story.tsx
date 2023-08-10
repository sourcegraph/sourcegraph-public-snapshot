import type { DecoratorFn, Story, Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesChangelogAlert } from './BatchChangesChangelogAlert'

const decorator: DecoratorFn = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangesChangelogAlert',
    decorators: [decorator],
    argTypes: {
        viewerIsAdmin: {
            name: 'Viewer is admin?',
            control: { type: 'boolean' },
            defaultValue: false,
        },
    },
}

export default config

export const Changelog: Story = args => (
    <WebStory>{() => <BatchChangesChangelogAlert viewerIsAdmin={args.viewerIsAdmin} />}</WebStory>
)
