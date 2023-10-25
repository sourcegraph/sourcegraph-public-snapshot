import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../components/WebStory'
import { IMPORTING_CHANGESETS, WORKSPACE_RESOLUTION_STATUS, WORKSPACES } from '../../../create/backend'
import {
    mockBatchChange,
    mockBatchSpec,
    mockBatchSpecImportingChangesets,
    mockBatchSpecWorkspaces,
    mockWorkspaceResolutionStatus,
    UNSTARTED_CONNECTION_MOCKS,
    UNSTARTED_WITH_CACHE_CONNECTION_MOCKS,
    LARGE_SUCCESS_CONNECTION_MOCKS,
    UNLICENSED_MOCK,
    LICENSED_MOCK,
} from '../../batch-spec.mock'
import { BatchSpecContextProvider } from '../../BatchSpecContext'

import { WorkspacesPreview } from './WorkspacesPreview'

const decorator: Decorator = story => (
    <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>
)

const config: Meta = {
    title: 'web/batches/batch-spec/edit/workspaces-preview/WorkspacesPreview',
    decorators: [decorator],
}

export default config

export const Unstarted: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={args.batchSpec ? mockBatchSpec() : mockBatchSpec({ originalInput: 'not-valid' })}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)
Unstarted.argTypes = {
    batchSpec: {
        name: 'Valid batch spec?',
        control: { type: 'boolean' },
    },
}
Unstarted.args = {
    batchSpec: true,
}

export const UnstartedWithCachedConnectionResult: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_WITH_CACHE_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={args.batchSpec ? mockBatchSpec() : mockBatchSpec({ originalInput: 'not-valid' })}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)
UnstartedWithCachedConnectionResult.argTypes = {
    batchSpec: {
        name: 'Valid batch spec?',
        control: { type: 'boolean' },
    },
}
UnstartedWithCachedConnectionResult.args = {
    batchSpec: true,
}

UnstartedWithCachedConnectionResult.storyName = 'unstarted, with cached connection result'

export const QueuedInProgress: StoryFn = args => {
    const inProgressResolution = mockWorkspaceResolutionStatus(args.inProgressResolution)

    const inProgressConnectionMocks = new WildcardMockLink([
        LICENSED_MOCK,
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
            {() => (
                <MockedTestProvider link={inProgressConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        batchSpec={args.batchSpec ? mockBatchSpec() : mockBatchSpec({ originalInput: 'not-valid' })}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
QueuedInProgress.argTypes = {
    inProgressResolution: {
        name: 'Status',
        control: {
            type: 'select',
            options: [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
        },
    },
    batchSpec: {
        control: { type: 'boolean' },
    },
}
QueuedInProgress.args = {
    inProgressResolution: BatchSpecWorkspaceResolutionState.QUEUED,
    batchSpec: true,
}

QueuedInProgress.storyName = 'queued/in progress'

export const QueuedInProgressWithCachedConnectionResult: StoryFn = args => {
    const inProgressResolution = mockWorkspaceResolutionStatus(args.inProgressResolution)

    const inProgressConnectionMocks = new WildcardMockLink([
        LICENSED_MOCK,
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
            {() => (
                <MockedTestProvider link={inProgressConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        batchSpec={mockBatchSpec()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
QueuedInProgressWithCachedConnectionResult.argTypes = {
    inProgressResolution: {
        name: 'Status',
        control: {
            type: 'select',
            options: [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
        },
    },
}
QueuedInProgressWithCachedConnectionResult.args = {
    inProgressResolution: BatchSpecWorkspaceResolutionState.QUEUED,
}

QueuedInProgressWithCachedConnectionResult.storyName = 'queued/in progress, with cached connection result'

export const FailedErrored: StoryFn = args => {
    const failedResolution = mockWorkspaceResolutionStatus(
        args.inProgressResolution,
        "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that."
    )

    const failedConnectionMocks = new WildcardMockLink([
        LICENSED_MOCK,
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
            {() => (
                <MockedTestProvider link={failedConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        batchSpec={mockBatchSpec()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
FailedErrored.argTypes = {
    inProgressResolution: {
        name: 'Status',
        control: {
            type: 'select',
            options: [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
        },
    },
}
FailedErrored.args = {
    inProgressResolution: BatchSpecWorkspaceResolutionState.FAILED,
}

FailedErrored.storyName = 'failed/errored'

export const FailedErroredWithCachedConnectionResult: StoryFn = args => {
    const failedResolution = mockWorkspaceResolutionStatus(
        args.inProgressResolution,
        "Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that."
    )

    const failedConnectionMocks = new WildcardMockLink([
        LICENSED_MOCK,
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
            {() => (
                <MockedTestProvider link={failedConnectionMocks}>
                    <BatchSpecContextProvider
                        batchChange={mockBatchChange()}
                        batchSpec={mockBatchSpec()}
                        refetchBatchChange={() => Promise.resolve()}
                    >
                        <WorkspacesPreview />
                    </BatchSpecContextProvider>
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
FailedErroredWithCachedConnectionResult.argTypes = {
    inProgressResolution: {
        name: 'Status',
        control: {
            type: 'select',
            options: [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
        },
    },
}
FailedErroredWithCachedConnectionResult.args = {
    inProgressResolution: BatchSpecWorkspaceResolutionState.FAILED,
}

FailedErroredWithCachedConnectionResult.storyName = 'failed/errored, with cached connection result'

export const Succeeded: StoryFn = () => (
    <WebStory>
        {() => (
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
                            noCache: false,
                        },
                    }}
                >
                    <WorkspacesPreview />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

export const CacheDisabled: StoryFn = () => (
    <WebStory>
        {() => (
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
                            noCache: true,
                        },
                    }}
                >
                    <WorkspacesPreview />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

export const ReadOnly: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink(UNSTARTED_WITH_CACHE_CONNECTION_MOCKS)}>
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={mockBatchSpec()}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview isReadOnly={true} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

ReadOnly.storyName = 'read-only'

export const SucceededWithScaleAlert: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink(LARGE_SUCCESS_CONNECTION_MOCKS)}>
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
                            noCache: false,
                        },
                    }}
                >
                    <WorkspacesPreview />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

SucceededWithScaleAlert.storyName = 'succeeded, with size alert'

export const ReadOnlyWithScaleAlert: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={new WildcardMockLink(LARGE_SUCCESS_CONNECTION_MOCKS)}>
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
)

ReadOnlyWithScaleAlert.storyName = 'read-only, with size alert'

export const UnstartedWithLicenseAlertConnectionResult: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                link={new WildcardMockLink([UNLICENSED_MOCK, ...UNSTARTED_WITH_CACHE_CONNECTION_MOCKS])}
            >
                <BatchSpecContextProvider
                    batchChange={mockBatchChange()}
                    batchSpec={mockBatchSpec()}
                    refetchBatchChange={() => Promise.resolve()}
                >
                    <WorkspacesPreview {...props} isReadOnly={false} />
                </BatchSpecContextProvider>
            </MockedTestProvider>
        )}
    </WebStory>
)

UnstartedWithLicenseAlertConnectionResult.storyName = 'unstarted, with license alert'

export const ReadOnlyWithLicenseAlert: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                link={new WildcardMockLink([UNLICENSED_MOCK, ...UNSTARTED_WITH_CACHE_CONNECTION_MOCKS])}
            >
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
)

ReadOnlyWithLicenseAlert.storyName = 'read-only, with license alert'
