import { boolean, select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../components/WebStory'
import {
    BatchSpecImportingChangesetsResult,
    BatchSpecWorkspacesPreviewResult,
    WorkspaceResolutionStatusResult,
} from '../../../../../graphql-operations'
import { IMPORTING_CHANGESETS, WORKSPACES, WORKSPACE_RESOLUTION_STATUS } from '../../../create/backend'
import { mockBatchChange, mockBatchSpec, mockImportingChangesets, mockWorkspaces } from '../../batch-spec.mock'
import { BatchSpecContextProvider } from '../../BatchSpecContext'

import { WorkspacesPreview } from './WorkspacesPreview'

const { add } = storiesOf('web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreview', module)
    .addDecorator(story => <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>)
    .addParameters({ chromatic: { disableSnapshot: true } })

const NODE_WITH_NO_WORKSPACES: BatchSpecWorkspacesPreviewResult = {
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: {
            __typename: 'BatchSpecWorkspaceResolution',
            workspaces: {
                __typename: 'BatchSpecWorkspaceConnection',
                totalCount: 0,
                pageInfo: { hasNextPage: false, endCursor: null },
                nodes: [],
            },
        },
    },
}

const NODE_WITH_NO_CHANGESETS: BatchSpecImportingChangesetsResult = {
    node: {
        __typename: 'BatchSpec',
        importingChangesets: {
            __typename: 'ChangesetSpecConnection',
            totalCount: 0,
            pageInfo: { hasNextPage: false, endCursor: null },
            nodes: [],
        },
    },
}

const UNSTARTED_RESOLUTION: WorkspaceResolutionStatusResult = {
    node: { __typename: 'BatchSpec', workspaceResolution: null },
}

const UNSTARTED_CONNECTION_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(WORKSPACES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: NODE_WITH_NO_WORKSPACES },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(IMPORTING_CHANGESETS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: NODE_WITH_NO_CHANGESETS },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: UNSTARTED_RESOLUTION },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

add('unstarted', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={UNSTARTED_CONNECTION_MOCKS}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange({
                        batchSpecs: {
                            __typename: 'BatchSpecConnection',
                            nodes: [
                                boolean('Valid batch spec?', true)
                                    ? mockBatchSpec()
                                    : mockBatchSpec({ originalInput: 'not-valid' }),
                            ],
                        },
                    })}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview {...props} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
))

const NODE_WITH_WORKSPACES: BatchSpecWorkspacesPreviewResult = {
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: {
            __typename: 'BatchSpecWorkspaceResolution',
            workspaces: {
                __typename: 'BatchSpecWorkspaceConnection',
                totalCount: 0,
                pageInfo: { hasNextPage: true, endCursor: 'cursor' },
                nodes: mockWorkspaces(50),
            },
        },
    },
}

const NODE_WITH_CHANGESETS: BatchSpecImportingChangesetsResult = {
    node: {
        __typename: 'BatchSpec',
        importingChangesets: {
            __typename: 'ChangesetSpecConnection',
            totalCount: 0,
            pageInfo: { hasNextPage: false, endCursor: null },
            nodes: mockImportingChangesets(10),
        },
    },
}

const COMPLETED_RESOLUTION: WorkspaceResolutionStatusResult = {
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: { state: BatchSpecWorkspaceResolutionState.COMPLETED, failureMessage: null },
    },
}

const UNSTARTED_WITH_CACHE_CONNECTION_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(WORKSPACES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: NODE_WITH_WORKSPACES },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(IMPORTING_CHANGESETS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: NODE_WITH_CHANGESETS },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: COMPLETED_RESOLUTION },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

add('unstarted, with cached connection result', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={UNSTARTED_WITH_CACHE_CONNECTION_MOCKS}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange({
                        batchSpecs: {
                            __typename: 'BatchSpecConnection',
                            nodes: [
                                boolean('Valid batch spec?', true)
                                    ? mockBatchSpec()
                                    : mockBatchSpec({ originalInput: 'not-valid' }),
                            ],
                        },
                    })}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview {...props} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
))

add('queued/in progress', () => {
    const inProgressResolution: WorkspaceResolutionStatusResult = {
        node: {
            __typename: 'BatchSpec',
            workspaceResolution: {
                state: select(
                    'Status',
                    [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
                    BatchSpecWorkspaceResolutionState.QUEUED
                ),
                failureMessage: null,
            },
        },
    }

    const inProgressConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_NO_WORKSPACES },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_NO_CHANGESETS },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: inProgressResolution },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={inProgressConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange({
                            batchSpecs: {
                                __typename: 'BatchSpecConnection',
                                nodes: [
                                    boolean('Valid batch spec?', true)
                                        ? mockBatchSpec()
                                        : mockBatchSpec({ originalInput: 'not-valid' }),
                                ],
                            },
                        })}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview {...props} />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('queued/in progress, with cached connection result', () => {
    const inProgressResolution: WorkspaceResolutionStatusResult = {
        node: {
            __typename: 'BatchSpec',
            workspaceResolution: {
                state: select(
                    'Status',
                    [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
                    BatchSpecWorkspaceResolutionState.QUEUED
                ),
                failureMessage: null,
            },
        },
    }

    const inProgressConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_WORKSPACES },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_CHANGESETS },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: inProgressResolution },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={inProgressConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview {...props} />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('failed/errored', () => {
    const failedResolution: WorkspaceResolutionStatusResult = {
        node: {
            __typename: 'BatchSpec',
            workspaceResolution: {
                state: select(
                    'Status',
                    [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
                    BatchSpecWorkspaceResolutionState.FAILED
                ),
                failureMessage:
                    "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that.",
            },
        },
    }

    const failedConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_NO_WORKSPACES },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_NO_CHANGESETS },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: failedResolution },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={failedConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview {...props} />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('failed/errored, with cached connection result', () => {
    const failedResolution: WorkspaceResolutionStatusResult = {
        node: {
            __typename: 'BatchSpec',
            workspaceResolution: {
                state: select(
                    'Status',
                    [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
                    BatchSpecWorkspaceResolutionState.FAILED
                ),
                failureMessage:
                    "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that.",
            },
        },
    }

    const failedConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_WORKSPACES },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: NODE_WITH_CHANGESETS },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: failedResolution },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={failedConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview {...props} />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
})
