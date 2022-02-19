import { storiesOf } from '@storybook/react'
import { upperFirst } from 'lodash'
import React from 'react'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState, BatchSpecState } from '../../../graphql-operations'

import { BatchChangeStatePill, BatchChangeStatePillProps } from './BatchChangeStatePill'

const { add } = storiesOf('web/batches/list', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const STATE_COMBINATIONS: BatchChangeStatePillProps[] = [
    { state: BatchChangeState.DRAFT },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.PENDING },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.QUEUED },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.PROCESSING },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.FAILED },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.COMPLETED },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.CANCELING },
    { state: BatchChangeState.DRAFT, latestExecutionState: BatchSpecState.CANCELED },
    // This state would only come from a batch change executed with src-cli.
    { state: BatchChangeState.OPEN },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.PENDING },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.QUEUED },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.PROCESSING },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.FAILED },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.COMPLETED },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.CANCELING },
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.CANCELED },
    { state: BatchChangeState.CLOSED },
    // If it's closed, we don't care about execution state, but let's just check one.
    { state: BatchChangeState.OPEN, latestExecutionState: BatchSpecState.PENDING },
]

add('BatchChangeStatePill', () => (
    <WebStory>
        {props => (
            <div className="d-flex flex-column align-items-start">
                {STATE_COMBINATIONS.map(({ state, latestExecutionState }) => (
                    <React.Fragment key={`${state}-${latestExecutionState || ''}`}>
                        <h3>
                            {upperFirst(state.toLowerCase())}
                            {latestExecutionState ? `, ${upperFirst(latestExecutionState.toLowerCase())}` : ''}
                        </h3>
                        <BatchChangeStatePill
                            className="mt-1 mb-3"
                            {...props}
                            state={state}
                            latestExecutionState={latestExecutionState}
                        />
                    </React.Fragment>
                ))}
            </div>
        )}
    </WebStory>
))
