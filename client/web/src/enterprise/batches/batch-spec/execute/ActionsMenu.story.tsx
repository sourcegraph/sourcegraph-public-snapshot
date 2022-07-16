import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import {
    COMPLETED_BATCH_SPEC,
    COMPLETED_WITH_ERRORS_BATCH_SPEC,
    EXECUTING_BATCH_SPEC,
    mockBatchChange,
} from '../batch-spec.mock'
import { BatchSpecContextProvider } from '../BatchSpecContext'

import { ActionsMenu } from './ActionsMenu'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/ActionsMenu',
    decorators: [decorator],
}

export default config

export const Executing: Story = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={EXECUTING_BATCH_SPEC}>
                <ActionsMenu />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

export const Failed: Story = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={COMPLETED_WITH_ERRORS_BATCH_SPEC}>
                <ActionsMenu />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

export const Completed: Story = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={COMPLETED_BATCH_SPEC}>
                <ActionsMenu />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

export const CompletedWithErrors: Story = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={COMPLETED_WITH_ERRORS_BATCH_SPEC}>
                <ActionsMenu />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

CompletedWithErrors.storyName = 'completed with errors'
