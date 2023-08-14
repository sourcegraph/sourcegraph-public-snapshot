import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { EditorFeedbackPanel } from './EditorFeedbackPanel'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/editor/EditorFeedbackPanel',
    decorators: [decorator],
    argTypes: {
        actions: {
            name: 'Actions',
            control: { type: 'text' },
            defaultValue: '',
        },
        execute: {
            name: 'Execute',
            control: { type: 'text' },
            defaultValue: '',
        },
        preview: {
            name: 'Preview',
            control: { type: 'text' },
            defaultValue: '',
        },
        codeUpdate: {
            name: 'Code Update',
            control: { type: 'text' },
            defaultValue: '',
        },
        codeValidation: {
            name: 'codeValidation',
            control: { type: 'text' },
            defaultValue: 'The entered spec is invalid:\n  * name must match pattern "^my-batch-change$"',
        },
    },
}

export default config

export const EditorFeedbackPanelStory: Story = args => (
    <WebStory>
        {props => (
            <EditorFeedbackPanel
                errors={{
                    actions: args.actions,
                    execute: args.execute,
                    preview: args.preview,
                    codeUpdate: args.codeUpdate,
                    codeValidation: args.codeValidation,
                }}
                {...props}
            />
        )}
    </WebStory>
)

EditorFeedbackPanelStory.storyName = 'EditorFeedbackPanel'
