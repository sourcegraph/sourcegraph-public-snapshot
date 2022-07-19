import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { MonacoBatchSpecEditor } from './MonacoBatchSpecEditor'

import sample from '../library/hello-world.batch.yaml'
import { boolean, withKnobs } from '@storybook/addon-knobs'
import { action } from '@storybook/addon-actions'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/editor/MonacoBatchSpecEditor',
    decorators: [decorator, withKnobs],
}

export default config

export const MonacoBatchSpecEditorStory: Story = () => (
    <WebStory>
        {props => (
            <MonacoBatchSpecEditor
                batchChangeName="hello-world"
                value={sample}
                readOnly={boolean('Read Only', false)}
                autoFocus={boolean('Auto Focus', false)}
                onChange={action('On Change')}
                {...props}
            />
        )}
    </WebStory>
)

MonacoBatchSpecEditorStory.storyName = 'MonacoBatchSpecEditor'
