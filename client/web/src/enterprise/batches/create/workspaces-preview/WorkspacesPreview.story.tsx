import { boolean, select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { MATCH_ANY_PARAMETERS, WildcardMockedResponse, WildcardMockLink } from 'wildcard-mock-link'

import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/apollo'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { WORKSPACES_AND_IMPORTING_CHANGESETS, WORKSPACE_RESOLUTION_STATUS } from '../backend'

import { WorkspacesPreview } from './WorkspacesPreview'
import { mockWorkspaceResolutionStatus, mockWorkspacesAndImportingChangesets } from './WorkspacesPreview.mock'

const { add } = storiesOf('web/batches/CreateBatchChangePage/WorkspacesPreview', module).addDecorator(story => (
    <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>
))

add('initial', () => (
    <WebStory>
        {props => (
            <WorkspacesPreview
                {...props}
                // batchSpecInput={text('Batch spec input', minimalSample)}
                previewDisabled={!boolean('Valid batch spec?', true)}
                preview={noop}
                batchSpecStale={false}
                excludeRepo={noop}
            />
        )}
    </WebStory>
))

add('first preview, loading', () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: mockWorkspaceResolutionStatus(
                    select(
                        'Status',
                        [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
                        BatchSpecWorkspaceResolutionState.QUEUED
                    )
                ),
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspacesPreview
                        {...props}
                        batchSpecID="fakelol"
                        currentPreviewRequestTime="1234"
                        previewDisabled={false}
                        preview={noop}
                        batchSpecStale={false}
                        excludeRepo={noop}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('first preview, error', () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: mockWorkspaceResolutionStatus(
                    select(
                        'Status',
                        [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
                        BatchSpecWorkspaceResolutionState.FAILED
                    ),
                    "Uh oh something bad happened and the workspace resolution failed! Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that"
                ),
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspacesPreview
                        {...props}
                        batchSpecID="fakelol"
                        currentPreviewRequestTime="1234"
                        previewDisabled={false}
                        preview={noop}
                        batchSpecStale={false}
                        excludeRepo={noop}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

const WORKSPACES_AND_IMPORTING_CHANGESETS_MOCK: WildcardMockedResponse = {
    request: {
        query: getDocumentNode(WORKSPACES_AND_IMPORTING_CHANGESETS),
        variables: MATCH_ANY_PARAMETERS,
    },
    result: {
        data: mockWorkspacesAndImportingChangesets(10, 2),
    },
    nMatches: Number.POSITIVE_INFINITY,
}

add('first preview, success', () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED),
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        WORKSPACES_AND_IMPORTING_CHANGESETS_MOCK,
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspacesPreview
                        {...props}
                        batchSpecID="fakelol"
                        currentPreviewRequestTime="1234"
                        previewDisabled={false}
                        preview={noop}
                        batchSpecStale={false}
                        excludeRepo={noop}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('first preview, stale', () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED),
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        WORKSPACES_AND_IMPORTING_CHANGESETS_MOCK,
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspacesPreview
                        {...props}
                        batchSpecID="fakelol"
                        currentPreviewRequestTime="1234"
                        previewDisabled={false}
                        preview={noop}
                        batchSpecStale={true}
                        excludeRepo={noop}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

// TODO: Add these stories once the workspaces preview list is kept visible on subsequent updates
// add('subsequent preview, loading', () => {})
