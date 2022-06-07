import { storiesOf } from '@storybook/react'
import { addMinutes } from 'date-fns'
import { MATCH_ANY_PARAMETERS, MockedResponses, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
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
    FAILED_BATCH_SPEC,
    mockBatchChange,
    mockWorkspaceResolutionStatus,
    mockWorkspaces,
} from '../batch-spec.mock'

import { BATCH_SPEC_WORKSPACES, FETCH_BATCH_SPEC_EXECUTION } from './backend'
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

add('executing', () => (
    <WebStory>
        {props => {
            const mock = EXECUTING_BATCH_SPEC

            // A true executing batch spec wouldn't have a finishedAt set, but
            // we need to have one so that Chromatic doesn't exhibit flakiness
            // based on how long it takes to actually take the snapshot, since
            // the timer in ExecuteBatchSpecPage is live in that case.
            mock.finishedAt = addMinutes(Date.parse(mock.startedAt!), 15).toISOString()

            return (
                <MockedTestProvider link={new WildcardMockLink(buildMocks({ ...EXECUTING_BATCH_SPEC }))}>
                    <ExecuteBatchSpecPage
                        {...props}
                        batchSpecID="spec1234"
                        batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                        authenticatedUser={mockAuthenticatedUser}
                        settingsCascade={SETTINGS_CASCADE}
                    />
                </MockedTestProvider>
            )
        }}
    </WebStory>
))

const FAILED_MOCKS = buildMocks(FAILED_BATCH_SPEC, { state: BatchSpecWorkspaceState.FAILED, failureMessage: 'Uh oh!' })

add('failed', () => {
    console.log(mockWorkspaces(50, { state: BatchSpecWorkspaceState.FAILED, failureMessage: 'Uh oh!' }))
    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={new WildcardMockLink(FAILED_MOCKS)}>
                    <ExecuteBatchSpecPage
                        {...props}
                        batchSpecID="spec1234"
                        batchChange={{ name: 'my-batch-change', namespace: 'user1234' }}
                        testContextState={{
                            errors: {
                                execute:
                                    "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that.",
                            },
                        }}
                        authenticatedUser={mockAuthenticatedUser}
                        settingsCascade={SETTINGS_CASCADE}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

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
