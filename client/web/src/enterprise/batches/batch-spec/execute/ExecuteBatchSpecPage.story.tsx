import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, type MockedResponses, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import type { AuthenticatedUser } from '../../../../auth'
import { WebStory } from '../../../../components/WebStory'
import {
    type BatchSpecExecutionFields,
    BatchSpecSource,
    BatchSpecWorkspaceResolutionState,
    BatchSpecWorkspaceState,
    type VisibleBatchSpecWorkspaceFields,
} from '../../../../graphql-operations'
import { GET_BATCH_CHANGE_TO_EDIT, WORKSPACE_RESOLUTION_STATUS } from '../../create/backend'
import {
    COMPLETED_BATCH_SPEC,
    COMPLETED_WITH_ERRORS_BATCH_SPEC,
    EXECUTING_BATCH_SPEC,
    mockBatchChange,
    mockFullBatchSpec,
    mockWorkspaceResolutionStatus,
    mockWorkspaces,
    PROCESSING_WORKSPACE,
} from '../batch-spec.mock'

import {
    type queryWorkspacesList as _queryWorkspacesList,
    BATCH_SPEC_WORKSPACE_BY_ID,
    FETCH_BATCH_SPEC_EXECUTION,
} from './backend'
import { ExecuteBatchSpecPage } from './ExecuteBatchSpecPage'

const decorator: Decorator = story => (
    <div className="p-3" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/batches/batch-spec/execute/ExecuteBatchSpecPage',

    decorators: [decorator],
}

export default config

const MOCK_ORGANIZATION = {
    __typename: 'Org',
    name: 'acme-corp',
    id: 'acme-corp-id',
}

const mockAuthenticatedUser = {
    __typename: 'User',
    username: 'alice',
    displayName: 'alice',
    id: 'b',
    organizations: {
        nodes: [MOCK_ORGANIZATION],
    },
} as AuthenticatedUser

const COMMON_MOCKS: MockedResponses = [
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
            query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED) },
        nMatches: Number.POSITIVE_INFINITY,
    },
]

const buildMocks = (batchSpec: BatchSpecExecutionFields): MockedResponses => [
    ...COMMON_MOCKS,
    {
        request: {
            query: getDocumentNode(FETCH_BATCH_SPEC_EXECUTION),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { node: batchSpec } },
        nMatches: Number.POSITIVE_INFINITY,
    },
]

const buildWorkspacesQuery =
    (workspaceFields?: Partial<VisibleBatchSpecWorkspaceFields>): typeof _queryWorkspacesList =>
    () =>
        of(mockWorkspaces(50, workspaceFields).node.workspaceResolution!.workspaces)

export const Executing: StoryFn = () => (
    <WebStory
        path="/users/:username/batch-changes/:batchChangeName/executions/:batchSpecID/*"
        initialEntries={['/users/my-username/batch-changes/my-batch-change/executions/spec1234']}
    >
        {props => (
            <MockedTestProvider link={new WildcardMockLink(buildMocks({ ...EXECUTING_BATCH_SPEC }))}>
                <ExecuteBatchSpecPage
                    {...props}
                    namespace={{ __typename: 'User', url: '', id: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery()}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const ExecuteWithAWorkspaceSelected: StoryFn = () => (
    <WebStory
        path="/users/:username/batch-changes/:batchChangeName/executions/:batchSpecID/*"
        initialEntries={[
            '/users/my-username/batch-changes/my-batch-change/executions/spec1234/execution/workspaces/workspace1234',
        ]}
    >
        {props => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        ...buildMocks({ ...EXECUTING_BATCH_SPEC }),
                        {
                            request: {
                                query: getDocumentNode(BATCH_SPEC_WORKSPACE_BY_ID),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: { data: { node: PROCESSING_WORKSPACE } },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <ExecuteBatchSpecPage
                    {...props}
                    namespace={{ __typename: 'User', url: '', id: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery()}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExecuteWithAWorkspaceSelected.storyName = 'executing, with a workspace selected'

const COMPLETED_MOCKS = buildMocks(COMPLETED_BATCH_SPEC)

export const Completed: StoryFn = () => (
    <WebStory initialEntries={['/my-batch-change/spec1234/execution']} path="/:batchChangeName/:batchSpecID/*">
        {props => (
            <MockedTestProvider link={new WildcardMockLink(COMPLETED_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    namespace={{ __typename: 'User', url: '', id: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery({ state: BatchSpecWorkspaceState.COMPLETED })}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

const COMPLETED_WITH_ERRORS_MOCKS = buildMocks(COMPLETED_WITH_ERRORS_BATCH_SPEC)

export const CompletedWithErrors: StoryFn = () => (
    <WebStory initialEntries={['/my-batch-change/spec1234/execution']} path="/:batchChangeName/:batchSpecID/*">
        {props => (
            <MockedTestProvider link={new WildcardMockLink(COMPLETED_WITH_ERRORS_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    namespace={{ __typename: 'User', url: '', id: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery({ state: BatchSpecWorkspaceState.FAILED })}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

CompletedWithErrors.storyName = 'completed with errors'

const LOCAL_MOCKS = buildMocks(mockFullBatchSpec({ source: BatchSpecSource.LOCAL }))

export const LocallyExecutedSpec: StoryFn = () => (
    <WebStory initialEntries={['/my-local-batch-change/spec1234/execution']} path="/:batchChangeName/:batchSpecID/*">
        {props => (
            <MockedTestProvider link={new WildcardMockLink(LOCAL_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    namespace={{ __typename: 'User', url: '', id: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

LocallyExecutedSpec.storyName = 'for a locally-executed spec'
