import type { StoryFn, Meta, Decorator } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'
import { WildcardMockLink, MATCH_ANY_PARAMETERS } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { BatchChangeState } from '../../../../graphql-operations'
import { CHANGESETS, type queryExternalChangesetWithFileDiffs } from '../backend'

import { BatchChangeChangesets } from './BatchChangeChangesets'
import { BATCH_CHANGE_CHANGESETS_RESULT, EMPTY_BATCH_CHANGE_CHANGESETS_RESULT } from './BatchChangeChangesets.mock'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangeChangesets',
    decorators: [decorator],
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

const mocks = new WildcardMockLink([
    {
        request: { query: getDocumentNode(CHANGESETS), variables: MATCH_ANY_PARAMETERS },
        result: { data: { node: BATCH_CHANGE_CHANGESETS_RESULT } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

const emptyMockData = new WildcardMockLink([
    {
        request: { query: getDocumentNode(CHANGESETS), variables: MATCH_ANY_PARAMETERS },
        result: { data: { node: EMPTY_BATCH_CHANGE_CHANGESETS_RESULT } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

const queryEmptyExternalChangesetWithFileDiffs: typeof queryExternalChangesetWithFileDiffs = ({
    externalChangeset,
}) => {
    switch (externalChangeset) {
        case 'somechangesetCLOSED':
        case 'somechangesetMERGED':
        case 'somechangesetDELETED': {
            return of({
                diff: null,
            })
        }
        default: {
            return of({
                diff: {
                    __typename: 'PreviewRepositoryComparison',
                    fileDiffs: {
                        nodes: [],
                        totalCount: 0,
                        pageInfo: {
                            endCursor: null,
                            hasNextPage: false,
                        },
                    },
                },
            })
        }
    }
}

export const ListOfChangesets: StoryFn = args => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks}>
                <BatchChangeChangesets
                    {...props}
                    refetchBatchChange={noop}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    batchChangeID="batchid"
                    viewerCanAdminister={args.viewerCanAdminister}
                    batchChangeState={BatchChangeState.OPEN}
                    isExecutionEnabled={false}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfChangesets.storyName = 'List of changesets'

export const ListOfExpandedChangesets: StoryFn = args => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks}>
                <BatchChangeChangesets
                    {...props}
                    refetchBatchChange={noop}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    batchChangeID="batchid"
                    viewerCanAdminister={args.viewerCanAdminister}
                    expandByDefault={true}
                    batchChangeState={BatchChangeState.OPEN}
                    isExecutionEnabled={false}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfExpandedChangesets.storyName = 'List of expanded changesets'

export const DraftWithoutChangesets: StoryFn = args => {
    const batchChangeState = args.batchChangeState

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={emptyMockData}>
                    <BatchChangeChangesets
                        {...props}
                        refetchBatchChange={noop}
                        queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                        batchChangeID="batchid"
                        viewerCanAdminister={true}
                        expandByDefault={true}
                        batchChangeState={batchChangeState as BatchChangeState}
                        isExecutionEnabled={args.isExecutionEnabled}
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
DraftWithoutChangesets.argTypes = {
    batchChangeState: {
        control: { type: 'select', options: Object.keys(BatchChangeState) },
    },
    isExecutionEnabled: {
        control: { type: 'boolean' },
    },
    viewerCanAdminister: {
        table: {
            disable: true,
        },
    },
}
DraftWithoutChangesets.args = {
    batchChangeState: BatchChangeState.DRAFT,
    isExecutionEnabled: true,
}

DraftWithoutChangesets.storyName = 'Draft without changesets'
