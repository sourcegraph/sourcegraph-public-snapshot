import { useState } from '@storybook/addons'
import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeCloseAlert } from './BatchChangeCloseAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/close/BatchChangeCloseAlert',
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

export const HasOpenChangesets: StoryFn = args => {
    const [closeChangesets, setCloseChangesets] = useState(false)
    return (
        <WebStory>
            {props => (
                <BatchChangeCloseAlert
                    {...props}
                    batchChangeID="change123"
                    batchChangeURL="/users/john/batch-changes/change123"
                    totalCount={args.totalCount}
                    closeChangesets={closeChangesets}
                    setCloseChangesets={setCloseChangesets}
                    viewerCanAdminister={args.viewerCanAdminister}
                    closeBatchChange={() => Promise.resolve()}
                />
            )}
        </WebStory>
    )
}
HasOpenChangesets.argTypes = {
    totalCount: {
        control: { type: 'number' },
    },
}
HasOpenChangesets.args = {
    totalCount: 10,
}

HasOpenChangesets.storyName = 'Has open changesets'

export const NoOpenChangesets: StoryFn = args => {
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
                    viewerCanAdminister={args.viewerCanAdminister}
                    closeBatchChange={() => Promise.resolve()}
                />
            )}
        </WebStory>
    )
}

NoOpenChangesets.storyName = 'No open changesets'
