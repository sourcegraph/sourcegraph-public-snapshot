import { storiesOf } from '@storybook/react'
import { addMinutes } from 'date-fns'
import { MemoryRouter } from 'react-router'
import { MATCH_ANY_PARAMETERS, MockedResponses, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { BatchSpecSource } from '@sourcegraph/shared/src/schema'
import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import {
    BatchSpecExecutionFields,
    BatchSpecWorkspaceResolutionState,
    BatchSpecWorkspaceState,
    VisibleBatchSpecWorkspaceFields,
} from '../../../../graphql-operations'
import { mockAuthenticatedUser } from '../../../code-monitoring/testing/util'
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

import { BATCH_SPEC_WORKSPACES, BATCH_SPEC_WORKSPACE_BY_ID, FETCH_BATCH_SPEC_EXECUTION } from './backend'
import { ExecuteBatchSpecPage } from './ExecuteBatchSpecPage'

const { add } = storiesOf('web/batches/batch-spec/execute/ExecuteBatchSpecPage', module)
    .addDecorator(story => (
        <div className="p-3" style={{ height: '95vh', width: '100%' }}>
            {story()}
        </div>
    ))
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

const FIXTURE_ORG: SettingsOrgSubject = {
    __typename: 'Org',
    name: 'sourcegraph',
    displayName: 'Sourcegraph',
    id: 'a',
    viewerCanAdminister: true,
}

const FIXTURE_USER: SettingsUserSubject = {
    __typename: 'User',
    username: 'alice',
    displayName: 'alice',
    id: 'b',
    viewerCanAdminister: true,
}

const SETTINGS_CASCADE = {
    ...EMPTY_SETTINGS_CASCADE,
    subjects: [
        { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
        { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
    ],
}

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

const buildMocks = (
    batchSpec: BatchSpecExecutionFields,
    workspaceFields?: Partial<VisibleBatchSpecWorkspaceFields>
): MockedResponses => [
    ...COMMON_MOCKS,
    {
        request: {
            query: getDocumentNode(BATCH_SPEC_WORKSPACES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockWorkspaces(50, workspaceFields) },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(FETCH_BATCH_SPEC_EXECUTION),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { node: batchSpec } },
        nMatches: Number.POSITIVE_INFINITY,
    },
]

// A true executing batch spec wouldn't have a finishedAt set, but we need to have one so
// that Chromatic doesn't exhibit flakiness based on how long it takes to actually take
// the snapshot, since the timer in ExecuteBatchSpecPage is live in that case.
const EXECUTING_BATCH_SPEC_WITH_END_TIME = {
    ...EXECUTING_BATCH_SPEC,
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    finishedAt: addMinutes(Date.parse(EXECUTING_BATCH_SPEC.startedAt!), 15).toISOString(),
}

add('executing', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(buildMocks({ ...EXECUTING_BATCH_SPEC_WITH_END_TIME }))}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

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

add('executing, with a workspace selected', () => (
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
                        settingsCascade={SETTINGS_CASCADE}
                        match={{
                            isExact: false,
                            params: { batchChangeName: 'my-batch-change', batchSpecID: 'spec1234' },
                            path: '/users/my-username/batch-changes/:batchChangeName/executions/:batchSpecID',
                            url: '/users/my-username/batch-changes/my-batch-change/executions/spec1234',
                        }}
                    />
                </MemoryRouter>
            </MockedTestProvider>
        )}
    </WebStory>
))

const COMPLETED_MOCKS = buildMocks(COMPLETED_BATCH_SPEC, { state: BatchSpecWorkspaceState.COMPLETED })

add('completed', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(COMPLETED_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

const COMPLETED_WITH_ERRORS_MOCKS = buildMocks(COMPLETED_WITH_ERRORS_BATCH_SPEC, {
    state: BatchSpecWorkspaceState.COMPLETED,
})

add('completed with errors', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(COMPLETED_WITH_ERRORS_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

const LOCAL_MOCKS = buildMocks(mockFullBatchSpec({ source: BatchSpecSource.LOCAL }))

add('for a locally-executed spec', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(LOCAL_MOCKS)}>
                <ExecuteBatchSpecPage
                    {...props}
                    batchSpecID="spec1234"
                    batchChange={{ name: 'my-local-batch-change', namespace: 'user1234' }}
                    authenticatedUser={mockAuthenticatedUser}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))
