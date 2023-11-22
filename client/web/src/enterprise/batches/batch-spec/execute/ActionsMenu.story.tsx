import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { EXECUTING_BATCH_SPEC, mockBatchChange } from '../batch-spec.mock'
import { BatchSpecContextProvider } from '../BatchSpecContext'

import { ActionsMenu, ActionsMenuMode } from './ActionsMenu'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/ActionsMenu',
    decorators: [decorator],
}

export default config

export const Preview: StoryFn = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={EXECUTING_BATCH_SPEC}>
                <ActionsMenu defaultMode={ActionsMenuMode.Preview} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

export const Actions: StoryFn = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={EXECUTING_BATCH_SPEC}>
                <ActionsMenu defaultMode={ActionsMenuMode.Actions} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

export const ActionsOnlyClose: StoryFn = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={EXECUTING_BATCH_SPEC}>
                <ActionsMenu defaultMode={ActionsMenuMode.ActionsOnlyClose} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

export const ActionsWithPreview: StoryFn = () => (
    <WebStory>
        {() => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={EXECUTING_BATCH_SPEC}>
                <ActionsMenu defaultMode={ActionsMenuMode.ActionsWithPreview} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)
