import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { MultiSelectContextProvider } from '../MultiSelectContext'

import { CreateUpdateBatchChangeAlert } from './CreateUpdateBatchChangeAlert'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/preview/CreateUpdateBatchChangeAlert',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}

export default config

export const Create: Story = () => (
    <WebStory>
        {props => (
            <CreateUpdateBatchChangeAlert
                {...props}
                specID="123"
                toBeArchived={18}
                batchChange={null}
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
            />
        )}
    </WebStory>
)

export const Update: Story = () => (
    <WebStory>
        {props => (
            <CreateUpdateBatchChangeAlert
                {...props}
                specID="123"
                toBeArchived={199}
                batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
            />
        )}
    </WebStory>
)

export const Disabled: Story = () => (
    <WebStory>
        {props => (
            <MultiSelectContextProvider initialSelected={['id1', 'id2']}>
                <CreateUpdateBatchChangeAlert
                    {...props}
                    specID="123"
                    toBeArchived={199}
                    batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                />
            </MultiSelectContextProvider>
        )}
    </WebStory>
)
