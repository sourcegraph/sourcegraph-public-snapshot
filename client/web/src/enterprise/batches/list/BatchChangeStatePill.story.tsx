import React from 'react'

import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { upperFirst } from 'lodash'

import { H3 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState, BatchSpecState } from '../../../graphql-operations'

import { BatchChangeStatePill, type BatchChangeStatePillProps } from './BatchChangeStatePill'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/list',
    decorators: [decorator],
}

export default config

const buildTestProps = (
    state: BatchChangeState,
    firstID: number,
    isApplied: boolean,
    latestExecutionState?: BatchSpecState
): BatchChangeStatePillProps => ({
    state,
    latestExecutionState,
    currentSpecID: `${firstID}`,
    latestSpecID: isApplied ? `${firstID}` : `${firstID + 1}`,
})

// Some of these state combinations shouldn't be possible (e.g. failed and applied), but
// it's easier to include them just in case.
const STATE_COMBINATIONS: BatchChangeStatePillProps[] = Object.values(BatchChangeState).flatMap((state, index) =>
    // Latest execution state will be undefined if the batch change was executed locally
    // with src-cli.
    [undefined, ...Object.values(BatchSpecState)].flatMap((executionState, innerIndex) => [
        // Add a version where the latest batch spec has been applied, and one where it
        // has not.
        buildTestProps(state, index * 100 + innerIndex, true, executionState),
        buildTestProps(state, index * 100 + 50 + innerIndex, false, executionState),
    ])
)

export const BatchChangeStatePillStory: StoryFn = () => (
    <WebStory>
        {props => (
            <div className="d-flex flex-column align-items-start">
                {STATE_COMBINATIONS.map(({ state, latestExecutionState, currentSpecID, latestSpecID }) => (
                    <React.Fragment key={`${state}-${latestExecutionState || ''}-${currentSpecID}-${latestSpecID}`}>
                        <H3>
                            {upperFirst(state.toLowerCase())}
                            {latestExecutionState ? `, ${upperFirst(latestExecutionState.toLowerCase())}` : ''}
                            {currentSpecID === latestSpecID ? '' : ' (latest is not applied)'}
                        </H3>
                        <BatchChangeStatePill
                            className="mt-1 mb-3"
                            {...props}
                            state={state}
                            latestExecutionState={latestExecutionState}
                            currentSpecID={currentSpecID}
                            latestSpecID={latestSpecID}
                        />
                    </React.Fragment>
                ))}
            </div>
        )}
    </WebStory>
)

BatchChangeStatePillStory.storyName = 'BatchChangeStatePill'
