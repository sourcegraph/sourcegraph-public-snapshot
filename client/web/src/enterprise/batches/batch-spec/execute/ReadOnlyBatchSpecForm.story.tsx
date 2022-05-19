import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { BatchSpecState } from '../../../../graphql-operations'
import { mockBatchChange, mockFullBatchSpec } from '../batch-spec.mock'
import { BatchSpecContextProvider } from '../BatchSpecContext'

import { ReadOnlyBatchSpecForm } from './ReadOnlyBatchSpecForm'

const { add } = storiesOf('web/batches/batch-spec/execute/ReadOnlyBatchSpecForm', module).addDecorator(story => (
    <div className="p-3 d-flex" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
))

add('while executing', () => (
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
))

add('after execution finishes', () => (
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
))
