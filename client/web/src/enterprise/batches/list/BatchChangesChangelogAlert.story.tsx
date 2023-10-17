import type { Decorator, StoryFn, Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesChangelogAlert } from './BatchChangesChangelogAlert'

const decorator: Decorator = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangesChangelogAlert',
    decorators: [decorator],
    argTypes: {
        viewerIsAdmin: {
            name: 'Viewer is admin?',
            control: { type: 'boolean' },
        },
    },
    args: {
        viewerIsAdmin: false,
    },
}

export default config

export const Changelog: StoryFn = args => (
    <WebStory>{() => <BatchChangesChangelogAlert viewerIsAdmin={args.viewerIsAdmin} />}</WebStory>
)
