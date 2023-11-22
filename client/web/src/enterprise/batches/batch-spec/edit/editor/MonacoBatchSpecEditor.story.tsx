import { action } from '@storybook/addon-actions'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import sample from '../library/hello-world.batch.yaml'

import { MonacoBatchSpecEditor } from './MonacoBatchSpecEditor'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/editor/MonacoBatchSpecEditor',
    decorators: [decorator],
    argTypes: {
        readOnly: {
            name: 'Read Only',
            control: { type: 'boolean' },
        },
        autoFocus: {
            name: 'Auto Focus',
            control: { type: 'boolean' },
        },
    },
    args: {
        readOnly: false,
        autoFocus: false,
    },
}

export default config

export const MonacoBatchSpecEditorStory: StoryFn = args => (
    <WebStory>
        {props => (
            <MonacoBatchSpecEditor
                batchChangeName="hello-world"
                value={sample}
                readOnly={args.readOnly}
                autoFocus={args.autoFocus}
                onChange={action('On Change')}
                {...props}
            />
        )}
    </WebStory>
)

MonacoBatchSpecEditorStory.storyName = 'MonacoBatchSpecEditor'
