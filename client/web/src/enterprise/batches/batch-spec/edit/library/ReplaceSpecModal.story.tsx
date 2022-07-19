import { action } from '@storybook/addon-actions'
import { text, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { ReplaceSpecModal } from './ReplaceSpecModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/library/ReplaceSpecModal',
    decorators: [decorator, withKnobs],
}

export default config

export const ReplaceSpecModalStory: Story = () => (
    <WebStory>
        {props => (
            <ReplaceSpecModal
                libraryItemName={text('Name', 'my-batch-change')}
                onCancel={action('On Cancel')}
                onConfirm={action('On Confirm')}
                {...props}
            />
        )}
    </WebStory>
)

ReplaceSpecModalStory.storyName = 'ReplaceSpecModal'
