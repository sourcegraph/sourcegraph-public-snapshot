import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { addDays } from 'date-fns'
import { of } from 'rxjs'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../../components/WebStory'

import type { queryBatchSpecs as _queryBatchSpecs } from './backend'
import { BatchSpecsPage } from './BatchSpecsPage'
import { NODES, successNode } from './testData'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/specs/BatchSpecsPage',
    decorators: [decorator],
    parameters: {},
}

export default config

const NOW = () => addDays(new Date(), 1)

const queryBatchSpecs: typeof _queryBatchSpecs = () =>
    of({
        __typename: 'BatchSpecConnection',
        totalCount: 47,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [...NODES, successNode('pid1'), successNode('pid2'), successNode('pid3')],
    })

const queryNoBatchSpecs: typeof _queryBatchSpecs = () =>
    of({
        __typename: 'BatchSpecConnection',
        totalCount: 0,
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        nodes: [],
    })

export const ListOfSpecs: StoryFn = () => (
    <WebStory>
        {props => (
            <BatchSpecsPage
                {...props}
                queryBatchSpecs={queryBatchSpecs}
                now={NOW}
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )}
    </WebStory>
)

ListOfSpecs.storyName = 'List of specs'

export const NoSpecs: StoryFn = () => (
    <WebStory>
        {props => (
            <BatchSpecsPage
                {...props}
                queryBatchSpecs={queryNoBatchSpecs}
                now={NOW}
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )}
    </WebStory>
)

NoSpecs.storyName = 'No specs'
