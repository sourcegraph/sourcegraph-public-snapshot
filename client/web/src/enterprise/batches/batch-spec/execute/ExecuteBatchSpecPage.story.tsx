import { DecoratorFn, Meta, Story } from '@storybook/react'
import { addMinutes } from 'date-fns'
import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, MockedResponses, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { AuthenticatedUser } from '../../../../auth'
import { WebStory } from '../../../../components/WebStory'
import {
    BatchSpecExecutionFields,
    BatchSpecSource,
    BatchSpecWorkspaceResolutionState,
    BatchSpecWorkspaceState,
    VisibleBatchSpecWorkspaceFields,
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
    BATCH_SPEC_WORKSPACE_BY_ID,
    FETCH_BATCH_SPEC_EXECUTION,
    queryWorkspacesList as _queryWorkspacesList,
} from './backend'
import { ExecuteBatchSpecPage } from './ExecuteBatchSpecPage'

const decorator: DecoratorFn = story => (
    <div className="p-3" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/batches/batch-spec/execute/ExecuteBatchSpecPage',

    decorators: [decorator],

    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

const MOCK_ORGANIZATION = {
    __typename: 'Org',
    name: 'acme-corp',
    displayName: 'ACME Corporation',
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
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    () =>
        of(mockWorkspaces(50, workspaceFields).node.workspaceResolution!.workspaces)

// A true executing batch spec wouldn't have a finishedAt set, but we need to have one so
// that Chromatic doesn't exhibit flakiness based on how long it takes to actually take
// the snapshot, since the timer in ExecuteBatchSpecPage is live in that case.
const EXECUTING_BATCH_SPEC_WITH_END_TIME = {
    ...EXECUTING_BATCH_SPEC,
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    finishedAt: addMinutes(Date.parse(EXECUTING_BATCH_SPEC.startedAt!), 15).toISOString(),
}

export const Executing: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(buildMocks({ ...EXECUTING_BATCH_SPEC_WITH_END_TIME }))}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery()}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

// A true processing workspace wouldn't have a finishedAt set, but we need to have one so
// that Chromatic doesn't exhibit flakiness based on how long it takes to actually take
// the snapshot, since the timer in the workspace details section is live in that case.
const PROCESSING_WORKSPACE_WITH_END_TIMES = {
    ...PROCESSING_WORKSPACE,
    /* eslint-disable @typescript-eslint/no-non-null-assertion */
    finishedAt: addMinutes(Date.parse(PROCESSING_WORKSPACE.startedAt!), 15).toISOString(),
    steps: [
        PROCESSING_WORKSPACE.steps[0]!,
        {
            ...PROCESSING_WORKSPACE.steps[1],
            startedAt: null,
        },
        PROCESSING_WORKSPACE.steps[2]!,
    ],
    /* eslint-enable @typescript-eslint/no-non-null-assertion */
}

export const ExecuteWithAWorkspaceSelected: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        ...buildMocks({ ...EXECUTING_BATCH_SPEC_WITH_END_TIME }),
                        {
                            request: {
                                query: getDocumentNode(BATCH_SPEC_WORKSPACE_BY_ID),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: { data: { node: PROCESSING_WORKSPACE_WITH_END_TIMES } },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <MemoryRouter
                    initialEntries={[
                        '/users/my-username/batch-changes/my-batch-change/executions/spec1234/execution/workspaces/workspace1234',
                    ]}
                >
                    <ExecuteBatchSpecPage
                        {...props}
                        batchSpecID="spec1234"
                        batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                        authenticatedUser={mockAuthenticatedUser}
                        match={{
                            isExact: false,
                            params: { batchChangeName: 'my-batch-change', batchSpecID: 'spec1234' },
                            path: '/users/my-username/batch-changes/:batchChangeName/executions/:batchSpecID',
                            url: '/users/my-username/batch-changes/my-batch-change/executions/spec1234',
                        }}
                        queryWorkspacesList={buildWorkspacesQuery()}
                    />
                </MemoryRouter>
            </MockedTestProvider>
        )}
    </WebStory>
)

ExecuteWithAWorkspaceSelected.storyName = 'executing, with a workspace selected'

const COMPLETED_MOCKS = buildMocks(COMPLETED_BATCH_SPEC)

export const Completed: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(COMPLETED_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery({ state: BatchSpecWorkspaceState.COMPLETED })}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

const COMPLETED_WITH_ERRORS_MOCKS = buildMocks(COMPLETED_WITH_ERRORS_BATCH_SPEC)

export const CompletedWithErrors: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(COMPLETED_WITH_ERRORS_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    queryWorkspacesList={buildWorkspacesQuery({ state: BatchSpecWorkspaceState.FAILED })}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

CompletedWithErrors.storyName = 'completed with errors'

const LOCAL_MOCKS = buildMocks(mockFullBatchSpec({ source: BatchSpecSource.LOCAL }))

export const LocallyExecutedSpec: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(LOCAL_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-local-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

LocallyExecutedSpec.storyName = 'for a locally-executed spec'
