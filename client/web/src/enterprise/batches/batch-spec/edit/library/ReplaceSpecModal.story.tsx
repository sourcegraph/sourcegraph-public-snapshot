import { action } from '@storybook/addon-actions'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { ReplaceSpecModal } from './ReplaceSpecModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/library/ReplaceSpecModal',
    decorators: [decorator],
    argTypes: {
        libraryItemName: {
            name: 'Name',
            control: { type: 'text' },
        },
    },
    args: {
        libraryItemName: 'my-batch-change',
    },
}

export default config

export const ReplaceSpecModalStory: StoryFn = args => (
    <WebStory>
        {props => (
            <ReplaceSpecModal
                libraryItemName={args.libraryItemName}
                onCancel={action('On Cancel')}
                onConfirm={action('On Confirm')}
                {...props}
            />
        )}
    </WebStory>
)

ReplaceSpecModalStory.storyName = 'ReplaceSpecModal'
