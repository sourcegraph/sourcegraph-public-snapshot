import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { LibraryPane } from './LibraryPane'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/LibraryPane',
    decorators: [decorator],
}

export default config

export const Editable: StoryFn = () => (
    <WebStory>
        {props => <LibraryPane {...props} name="my-batch-change" onReplaceItem={() => alert('batch spec replaced!')} />}
    </WebStory>
)

export const ReadOnly: StoryFn = () => (
    <WebStory>{props => <LibraryPane {...props} name="my-batch-change" isReadOnly={true} />}</WebStory>
)

ReadOnly.storyName = 'read-only'
