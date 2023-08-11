import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { LibraryPane } from './LibraryPane'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/LibraryPane',
    decorators: [decorator],
}

export default config

export const Editable: Story = () => (
    <WebStory>
        {props => <LibraryPane {...props} name="my-batch-change" onReplaceItem={() => alert('batch spec replaced!')} />}
    </WebStory>
)

export const ReadOnly: Story = () => (
    <WebStory>{props => <LibraryPane {...props} name="my-batch-change" isReadOnly={true} />}</WebStory>
)

ReadOnly.storyName = 'read-only'
