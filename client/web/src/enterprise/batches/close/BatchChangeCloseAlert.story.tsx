import { boolean, number } from '@storybook/addon-knobs'
import { useState } from '@storybook/addons'
import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeCloseAlert } from './BatchChangeCloseAlert'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/close/BatchChangeCloseAlert',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}

export default config

export const HasOpenChangesets: Story = () => {
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
}

HasOpenChangesets.storyName = 'Has open changesets'

export const NoOpenChangesets: Story = () => {
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
}

NoOpenChangesets.storyName = 'No open changesets'
