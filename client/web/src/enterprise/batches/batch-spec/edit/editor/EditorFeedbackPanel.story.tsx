import { text, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { EditorFeedbackPanel } from './EditorFeedbackPanel'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/editor/EditorFeedbackPanel',
    decorators: [decorator, withKnobs],
}

export default config

export const EditorFeedbackPanelStory: Story = () => (
    <WebStory>
        {props => (
            <EditorFeedbackPanel
                errors={{
                    actions: text('Actions', ''),
                    execute: text('Execute', ''),
                    preview: text('Preview', ''),
                    codeUpdate: text('Code Update', ''),
                    codeValidation: text(
                        'Validation',
                        'The entered spec is invalid:\n  * name must match pattern "^my-batch-change$"'
                    ),
                }}
                {...props}
            />
        )}
    </WebStory>
)

EditorFeedbackPanelStory.storyName = 'EditorFeedbackPanel'
