import type { Decorator, StoryFn, Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { MultiSelectContextProvider } from '../MultiSelectContext'

import { CreateUpdateBatchChangeAlert } from './CreateUpdateBatchChangeAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/preview/CreateUpdateBatchChangeAlert',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
    argTypes: {
        viewerCanAdminister: {
            control: { type: 'boolean' },
        },
    },
    args: {
        viewerCanAdminister: true,
    },
}

export default config

export const Create: StoryFn = args => (
    <WebStory>
        {props => (
            <CreateUpdateBatchChangeAlert
                {...props}
                specID="123"
                toBeArchived={18}
                batchChange={null}
                viewerCanAdminister={args.viewerCanAdminister}
            />
        )}
    </WebStory>
)

export const Update: StoryFn = args => (
    <WebStory>
        {props => (
            <CreateUpdateBatchChangeAlert
                {...props}
                specID="123"
                toBeArchived={199}
                batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                viewerCanAdminister={args.viewerCanAdminister}
            />
        )}
    </WebStory>
)

export const Disabled: StoryFn = args => (
    <WebStory>
        {props => (
            <MultiSelectContextProvider initialSelected={['id1', 'id2']}>
                <CreateUpdateBatchChangeAlert
                    {...props}
                    specID="123"
                    toBeArchived={199}
                    batchChange={{ id: '123', name: 'awesome-batch-change', url: 'http://test.test/awesome' }}
                    viewerCanAdminister={args.viewerCanAdminister}
                />
            </MultiSelectContextProvider>
        )}
    </WebStory>
)
