import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { EditorFeedbackPanel } from './EditorFeedbackPanel'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/editor/EditorFeedbackPanel',
    decorators: [decorator],
    argTypes: {
        actions: {
            name: 'Actions',
            control: { type: 'text' },
        },
        execute: {
            name: 'Execute',
            control: { type: 'text' },
        },
        preview: {
            name: 'Preview',
            control: { type: 'text' },
        },
        codeUpdate: {
            name: 'Code Update',
            control: { type: 'text' },
        },
        codeValidation: {
            name: 'codeValidation',
            control: { type: 'text' },
        },
    },
    args: {
        actions: '',
        execute: '',
        preview: '',
        codeUpdate: '',
        codeValidation: 'The entered spec is invalid:\n  * name must match pattern "^my-batch-change$"',
    },
}

export default config

export const EditorFeedbackPanelStory: StoryFn = args => (
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
