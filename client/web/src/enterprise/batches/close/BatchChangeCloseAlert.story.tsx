import { boolean, number } from '@storybook/addon-knobs'
import { useState } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeCloseAlert } from './BatchChangeCloseAlert'

const { add } = storiesOf('web/batches/close/BatchChangeCloseAlert', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Has open changesets', () => {
    const [closeChangesets, setCloseChangesets] = useState(false)
    const totalCount = number('totalCount', 10)
    return (
        <WebStory>
            {props => (
                <BatchChangeCloseAlert
                    {...props}
                    batchChangeID="change123"
                    batchChangeURL="/users/john/batch-changes/change123"
                    totalCount={totalCount}
                    closeChangesets={closeChangesets}
                    setCloseChangesets={setCloseChangesets}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    closeBatchChange={() => Promise.resolve()}
                />
            )}
        </WebStory>
    )
})
add('No open changesets', () => {
    const [closeChangesets, setCloseChangesets] = useState(false)
    return (
        <WebStory>
            {props => (
                <BatchChangeCloseAlert
                    {...props}
                    batchChangeID="change123"
                    batchChangeURL="/users/john/batch-changes/change123"
                    totalCount={0}
                    closeChangesets={closeChangesets}
                    setCloseChangesets={setCloseChangesets}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    closeBatchChange={() => Promise.resolve()}
                />
            )}
        </WebStory>
    )
})
