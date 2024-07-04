import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../components/WebStory'
import { BatchSpecSource, BatchSpecWorkspaceResolutionState } from '../../../../../graphql-operations'
import { GET_BATCH_CHANGE_TO_EDIT, WORKSPACE_RESOLUTION_STATUS } from '../../../create/backend'
import {
    mockBatchChange,
    mockFullBatchSpec,
    mockWorkspace,
    mockWorkspaceResolutionStatus,
    mockWorkspaces,
} from '../../batch-spec.mock'
import { BatchSpecContextProvider } from '../../BatchSpecContext'
import {
    BATCH_SPEC_WORKSPACE_BY_ID,
    FETCH_BATCH_SPEC_EXECUTION,
    type queryWorkspacesList as _queryWorkspacesList,
} from '../backend'

import { ExecutionWorkspaces } from './ExecutionWorkspaces'

const decorator: Decorator = story => (
    <div className="p-3 d-flex" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/batches/batch-spec/execute/ExecutionWorkspaces',
    decorators: [decorator],
}

export default config

const MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: mockBatchChange() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(FETCH_BATCH_SPEC_EXECUTION),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { node: mockFullBatchSpec() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED) },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(BATCH_SPEC_WORKSPACE_BY_ID),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { node: mockWorkspace() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

const queryWorkspacesList: typeof _queryWorkspacesList = () =>
    of(mockWorkspaces(50).node.workspaceResolution!.workspaces)

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

export const List: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={MOCKS}>
                <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={mockFullBatchSpec()}>
                    <ExecutionWorkspaces
                        queryWorkspacesList={queryWorkspacesList}
                        {...props}
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

export const WorkspaceSelected: StoryFn = () => (
    <WebStory path="/:workspaceID" initialEntries={['/workspace2']}>
        {props => (
            <MockedTestProvider link={MOCKS}>
                <BatchSpecContextProvider batchChange={mockBatchChange()} batchSpec={mockFullBatchSpec()}>
                    <ExecutionWorkspaces
                        {...props}
                        queryWorkspacesList={queryWorkspacesList}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

WorkspaceSelected.storyName = 'with workspace selected'

export const LocallyExecutedSpec: StoryFn = () => (
    <WebStory path="/:workspaceID" initialEntries={['/"spec1234"']}>
        {props => (
            <MockedTestProvider link={MOCKS}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={mockFullBatchSpec({ source: BatchSpecSource.LOCAL })}
                >
                    <div className="container">
                        <ExecutionWorkspaces
                            {...props}
                            queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                            telemetryRecorder={noOpTelemetryRecorder}
                        />
                    </div>
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

LocallyExecutedSpec.storyName = 'for a locally-executed spec'
