import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import {
    COMPLETED_BATCH_SPEC,
    COMPLETED_WITH_ERRORS_BATCH_SPEC,
    EXECUTING_BATCH_SPEC,
    FAILED_BATCH_SPEC,
    mockBatchChange,
} from '../batch-spec.mock'
import { BatchSpecContextProvider } from '../BatchSpecContext'

import { ActionsMenu } from './ActionsMenu'

const { add } = storiesOf('web/batches/batch-spec/execute/ActionsMenu', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('executing', () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={EXECUTING_BATCH_SPEC}>
                <ActionsMenu {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
))

add('failed', () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={FAILED_BATCH_SPEC}>
                <ActionsMenu {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
))

add('completed', () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={COMPLETED_BATCH_SPEC}>
                <ActionsMenu {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
))

add('completed with errors', () => (
    <WebStory>
        {props => (
            <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={COMPLETED_WITH_ERRORS_BATCH_SPEC}>
                <ActionsMenu {...props} />
            </BatchSpecContextProvider>
        )}
    </WebStory>
))
