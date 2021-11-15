import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'
import { MultiSelectContextProvider } from '../MultiSelectContext'

import { CreateUpdateBatchChangeAlert } from './CreateUpdateBatchChangeAlert'

const { add } = storiesOf('web/batches/preview/CreateUpdateBatchChangeAlert', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Create', () => (
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
))
add('Update', () => (
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
))
add('Disabled', () => (
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
))
