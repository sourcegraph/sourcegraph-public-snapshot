import { boolean, select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../components/WebStory'
import { IMPORTING_CHANGESETS, WORKSPACES, WORKSPACE_RESOLUTION_STATUS } from '../../../create/backend'
import {
    mockBatchChange,
    mockBatchSpec,
    mockBatchSpecImportingChangesets,
    mockBatchSpecWorkspaces,
    mockWorkspaceResolutionStatus,
    UNSTARTED_CONNECTION_MOCKS,
    UNSTARTED_WITH_CACHE_CONNECTION_MOCKS,
} from '../../batch-spec.mock'
import { BatchSpecContextProvider } from '../../BatchSpecContext'

import { WorkspacesPreview } from './WorkspacesPreview'

const { add } = storiesOf(
    'web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreview',
    module
).addDecorator(story => <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>)

add('unstarted', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={
                        boolean('Valid batch spec?', true)
                            ? mockBatchSpec()
                            : mockBatchSpec({ originalInput: 'not-valid' })
                    }
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview {...props} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
))

add('unstarted, with cached connection result', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_WITH_CACHE_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={
                        boolean('Valid batch spec?', true)
                            ? mockBatchSpec()
                            : mockBatchSpec({ originalInput: 'not-valid' })
                    }
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview {...props} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
))

add('queued/in progress', () => {
    const inProgressResolution = mockWorkspaceResolutionStatus(
        select(
            'Status',
            [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
            BatchSpecWorkspaceResolutionState.QUEUED
        )
    )

    const inProgressConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecWorkspaces(0) },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecImportingChangesets(0) },
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
                        batchSpec={
                            boolean('Valid batch spec?', true)
                                ? mockBatchSpec()
                                : mockBatchSpec({ originalInput: 'not-valid' })
                        }
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
    const inProgressResolution = mockWorkspaceResolutionStatus(
        select(
            'Status',
            [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
            BatchSpecWorkspaceResolutionState.QUEUED
        )
    )

    const inProgressConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecWorkspaces(50) },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecImportingChangesets(20) },
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
                        batchSpec={mockBatchSpec()}
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
    const failedResolution = mockWorkspaceResolutionStatus(
        select(
            'Status',
            [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
            BatchSpecWorkspaceResolutionState.FAILED
        ),
        "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that."
    )

    const failedConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecWorkspaces(0) },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecImportingChangesets(0) },
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
                        batchSpec={mockBatchSpec()}
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
    const failedResolution = mockWorkspaceResolutionStatus(
        select(
            'Status',
            [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
            BatchSpecWorkspaceResolutionState.FAILED
        ),
        "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that."
    )

    const failedConnectionMocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACES),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecWorkspaces(50) },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(IMPORTING_CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: mockBatchSpecImportingChangesets(20) },
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
                        batchSpec={mockBatchSpec()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview {...props} />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('succeeded', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_WITH_CACHE_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={mockBatchSpec()}
                    refetchBatchChange={() => Promise.resolve()}
                    testState={{
                        workspacesPreview: {
                            hasPreviewed: true,
                            resolutionState: BatchSpecWorkspaceResolutionState.COMPLETED,
                            preview: () => Promise.resolve(),
                            cancel: noop,
                            isInProgress: false,
                            clearError: noop,
                            setFilters: noop,
                            isPreviewDisabled: false,
                        },
                    }}
                >
                    <WorkspacesPreview {...props} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
))

add('read-only', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_WITH_CACHE_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={mockBatchSpec()}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview {...props} isReadOnly={true} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
))
