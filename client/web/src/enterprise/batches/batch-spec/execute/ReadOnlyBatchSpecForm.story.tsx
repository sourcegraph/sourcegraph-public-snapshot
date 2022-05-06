import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { BatchSpecState } from '../../../../graphql-operations'
import { mockBatchChange, mockBatchSpec } from '../batch-spec.mock'

import { ReadOnlyBatchSpecForm } from './ReadOnlyBatchSpecForm'

const { add } = storiesOf('web/batches/batch-spec/execute/ReadOnlyBatchSpecForm', module)
    .addDecorator(story => (
        <div className="p-3 d-flex" style={{ height: '95vh', width: '100%' }}>
            {story()}
        </div>
    ))
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

add('while executing', () => (
    <WebStory>
        {props => (
            <ReadOnlyBatchSpecForm
                {...props}
                batchChange={mockBatchChange()}
                originalInput={mockBatchSpec().originalInput}
                executionState={select(
                    'batch spec state',
                    [BatchSpecState.PROCESSING, BatchSpecState.QUEUED],
                    BatchSpecState.PROCESSING
                )}
            />
        )}
    </WebStory>
))

add('after execution finishes', () => (
    <WebStory>
        {props => (
            <ReadOnlyBatchSpecForm
                {...props}
                batchChange={mockBatchChange()}
                originalInput={mockBatchSpec().originalInput}
                executionState={select(
                    'batch spec state',
                    [
                        BatchSpecState.CANCELED,
                        BatchSpecState.CANCELING,
                        BatchSpecState.COMPLETED,
                        BatchSpecState.FAILED,
                        BatchSpecState.PENDING,
                    ],
                    BatchSpecState.COMPLETED
                )}
            />
        )}
    </WebStory>
))
