import { select } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BatchSpecSource } from '@sourcegraph/shared/src/schema'

import { WebStory } from '../../../../components/WebStory'
import { BatchSpecState } from '../../../../graphql-operations'
import { mockBatchChange, mockFullBatchSpec } from '../batch-spec.mock'
import { BatchSpecContextProvider } from '../BatchSpecContext'

import { ReadOnlyBatchSpecForm } from './ReadOnlyBatchSpecForm'

const decorator: DecoratorFn = story => (
    <div className="p-3 d-flex" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
)
const config: Meta = {
    title: 'web/batches/batch-spec/execute/ReadOnlyBatchSpecForm',
    decorators: [decorator],
}

export default config

export const Executing: Story = () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider
                batchChange={mockBatchChange()}
                batchSpec={mockFullBatchSpec({
                    state: select(
                        'batch spec state',
                        [BatchSpecState.PROCESSING, BatchSpecState.QUEUED],
                        BatchSpecState.PROCESSING
                    ),
                })}
            >
                <ReadOnlyBatchSpecForm {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

Executing.storyName = 'while executing'

export const ExecutionFinished: Story = () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider
                batchChange={mockBatchChange()}
                batchSpec={mockFullBatchSpec({
                    state: select(
                        'batch spec state',
                        [
                            BatchSpecState.CANCELED,
                            BatchSpecState.CANCELING,
                            BatchSpecState.COMPLETED,
                            BatchSpecState.FAILED,
                            BatchSpecState.PENDING,
                        ],
                        BatchSpecState.COMPLETED
                    ),
                })}
            >
                <ReadOnlyBatchSpecForm {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

ExecutionFinished.storyName = 'after execution finishes'

export const LocallyExecutedSpec: Story = () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider
                batchChange={mockBatchChange()}
                batchSpec={mockFullBatchSpec({ source: BatchSpecSource.LOCAL })}
            >
                <ReadOnlyBatchSpecForm {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
)

LocallyExecutedSpec.storyName = 'for a locally-executed spec'
