import { boolean, select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { UseConnectionResult } from '../../../../components/FilteredConnection/hooks/useConnection'
import { PreviewBatchSpecWorkspaceFields } from '../../../../graphql-operations'

import { ImportingChangesetFields } from './useImportingChangesets'
import { WorkspacesPreview } from './WorkspacesPreview'
import { mockImportingChangesets, mockWorkspaces } from './WorkspacesPreview.mock'

const { add } = storiesOf('web/batches/CreateBatchChangePage/WorkspacesPreview', module)
    .addDecorator(story => <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>)
    .addParameters({ chromatic: { disableSnapshots: true } })

const EMPTY_RESOLUTION_CONNECTION: UseConnectionResult<PreviewBatchSpecWorkspaceFields> = {
    connection: {
        pageInfo: {
            hasNextPage: false,
        },
        nodes: [],
    },
    fetchMore: noop,
    refetchAll: noop,
    loading: false,
    hasNextPage: false,
    startPolling: noop,
    stopPolling: noop,
}

const EMPTY_CHANGESETS_CONNECTION: UseConnectionResult<ImportingChangesetFields> = {
    connection: {
        pageInfo: {
            hasNextPage: false,
        },
        nodes: [],
    },
    fetchMore: noop,
    refetchAll: noop,
    loading: false,
    hasNextPage: false,
    startPolling: noop,
    stopPolling: noop,
}

const FULL_RESOLUTION_CONNECTION: UseConnectionResult<PreviewBatchSpecWorkspaceFields> = {
    connection: {
        pageInfo: {
            hasNextPage: true,
        },
        nodes: mockWorkspaces(50),
    },
    fetchMore: noop,
    refetchAll: noop,
    loading: false,
    hasNextPage: true,
    startPolling: noop,
    stopPolling: noop,
}

const FULL_CHANGESETS_CONNECTION: UseConnectionResult<ImportingChangesetFields> = {
    connection: {
        pageInfo: {
            hasNextPage: false,
        },
        nodes: mockImportingChangesets(10),
    },
    fetchMore: noop,
    refetchAll: noop,
    loading: false,
    hasNextPage: false,
    startPolling: noop,
    stopPolling: noop,
}

add('unstarted', () => {
    const hasExistingPreview = boolean('Has existing preview', false)
    const workspacesConnection = hasExistingPreview ? FULL_RESOLUTION_CONNECTION : EMPTY_RESOLUTION_CONNECTION

    return (
        <WebStory>
            {props => (
                <WorkspacesPreview
                    {...props}
                    isWorkspacesPreviewInProgress={false}
                    cancel={noop}
                    resolutionState="UNSTARTED"
                    workspacesConnection={workspacesConnection}
                    importingChangesetsConnection={EMPTY_CHANGESETS_CONNECTION}
                    hasPreviewed={false}
                    previewDisabled={!boolean('Valid batch spec?', true)}
                    preview={noop}
                    batchSpecStale={false}
                    excludeRepo={noop}
                    setFilters={noop}
                />
            )}
        </WebStory>
    )
})

add('request in flight', () => {
    const hasExistingPreview = boolean('Has existing preview', false)
    const workspacesConnection = hasExistingPreview ? FULL_RESOLUTION_CONNECTION : EMPTY_RESOLUTION_CONNECTION

    return (
        <WebStory>
            {props => (
                <WorkspacesPreview
                    {...props}
                    isWorkspacesPreviewInProgress={true}
                    cancel={noop}
                    resolutionState="REQUESTED"
                    workspacesConnection={workspacesConnection}
                    importingChangesetsConnection={EMPTY_CHANGESETS_CONNECTION}
                    hasPreviewed={false}
                    previewDisabled={!boolean('Valid batch spec?', true)}
                    preview={noop}
                    batchSpecStale={false}
                    excludeRepo={noop}
                    setFilters={noop}
                />
            )}
        </WebStory>
    )
})

add('queued/in progress', () => {
    const hasExistingPreview = boolean('Has existing preview', false)
    const workspacesConnection = hasExistingPreview ? FULL_RESOLUTION_CONNECTION : EMPTY_RESOLUTION_CONNECTION

    return (
        <WebStory>
            {props => (
                <WorkspacesPreview
                    {...props}
                    isWorkspacesPreviewInProgress={true}
                    cancel={noop}
                    resolutionState={select(
                        'Status',
                        [BatchSpecWorkspaceResolutionState.QUEUED, BatchSpecWorkspaceResolutionState.PROCESSING],
                        BatchSpecWorkspaceResolutionState.QUEUED
                    )}
                    workspacesConnection={workspacesConnection}
                    importingChangesetsConnection={EMPTY_CHANGESETS_CONNECTION}
                    hasPreviewed={false}
                    previewDisabled={!boolean('Valid batch spec?', true)}
                    preview={noop}
                    batchSpecStale={false}
                    excludeRepo={noop}
                    setFilters={noop}
                />
            )}
        </WebStory>
    )
})

add('canceled', () => {
    const hasExistingPreview = boolean('Has existing preview', false)
    const workspacesConnection = hasExistingPreview ? FULL_RESOLUTION_CONNECTION : EMPTY_RESOLUTION_CONNECTION

    return (
        <WebStory>
            {props => (
                <WorkspacesPreview
                    {...props}
                    isWorkspacesPreviewInProgress={false}
                    cancel={noop}
                    resolutionState="CANCELED"
                    workspacesConnection={workspacesConnection}
                    importingChangesetsConnection={EMPTY_CHANGESETS_CONNECTION}
                    hasPreviewed={false}
                    previewDisabled={!boolean('Valid batch spec?', true)}
                    preview={noop}
                    batchSpecStale={false}
                    excludeRepo={noop}
                    setFilters={noop}
                />
            )}
        </WebStory>
    )
})

add('failed/errored', () => {
    const hasExistingPreview = boolean('Has existing preview', false)
    const workspacesConnection = hasExistingPreview ? FULL_RESOLUTION_CONNECTION : EMPTY_RESOLUTION_CONNECTION

    return (
        <WebStory>
            {props => (
                <WorkspacesPreview
                    {...props}
                    isWorkspacesPreviewInProgress={false}
                    cancel={noop}
                    resolutionState={select(
                        'Status',
                        [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
                        BatchSpecWorkspaceResolutionState.FAILED
                    )}
                    workspacesConnection={workspacesConnection}
                    importingChangesetsConnection={EMPTY_CHANGESETS_CONNECTION}
                    hasPreviewed={false}
                    previewDisabled={!boolean('Valid batch spec?', true)}
                    preview={noop}
                    batchSpecStale={false}
                    excludeRepo={noop}
                    setFilters={noop}
                    error="Oh no something went wrong. This is a longer error message to demonstrate how this might take up a decent portion of screen real estate but hopefully it's still helpful information so it's worth the cost. Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that."
                />
            )}
        </WebStory>
    )
})

add('success', () => (
    <WebStory>
        {props => (
            <WorkspacesPreview
                {...props}
                isWorkspacesPreviewInProgress={false}
                cancel={noop}
                resolutionState={BatchSpecWorkspaceResolutionState.COMPLETED}
                workspacesConnection={FULL_RESOLUTION_CONNECTION}
                importingChangesetsConnection={FULL_CHANGESETS_CONNECTION}
                hasPreviewed={true}
                previewDisabled={!boolean('Valid batch spec?', true)}
                preview={noop}
                batchSpecStale={boolean('Batch spec stale?', false)}
                excludeRepo={noop}
                setFilters={noop}
            />
        )}
    </WebStory>
))
