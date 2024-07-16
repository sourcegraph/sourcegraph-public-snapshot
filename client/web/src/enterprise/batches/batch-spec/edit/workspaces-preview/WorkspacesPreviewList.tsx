import React from 'react'

import type { Connection } from '../../../../../components/FilteredConnection'
import type { UseShowMorePaginationResult } from '../../../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../../../components/FilteredConnection/ui'
import type {
    BatchSpecWorkspacesPreviewResult,
    PreviewHiddenBatchSpecWorkspaceFields,
    PreviewVisibleBatchSpecWorkspaceFields,
} from '../../../../../graphql-operations'

import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

interface WorkspacesPreviewListProps {
    /** The current workspaces preview connection result used to render the list. */
    workspacesConnection: UseShowMorePaginationResult<
        BatchSpecWorkspacesPreviewResult,
        PreviewHiddenBatchSpecWorkspaceFields | PreviewVisibleBatchSpecWorkspaceFields
    >
    /**
     * Whether or not the workspaces in this list are up-to-date with the current batch
     * spec input YAML in the editor.
     */
    isStale: boolean
    /**
     * Function to automatically update repo query of input batch spec YAML to exclude the
     * provided repo + branch.
     */
    excludeRepo: (repo: string, branch: string) => void
    /**
     * Whether or not the preview list should prefer the `cached` list over the data
     * actively in the connection.
     */
    showCached: boolean
    /**
     * We "cache" the last results of the workspaces preview so that we can continue to
     * show them in the list while the next workspaces resolution is still in progress. We
     * have to do this outside of Apollo Client because we continue to requery the
     * orkspaces preview while the resolution job is still in progress, and so the results
     * will come up empty and overwrite the previous results in the Apollo Client cache
     * while this is happening. If data is availabled in `cached` and `showCached` is
     * true, it will be used over the data in the connnection.
     */
    cached?: Connection<PreviewHiddenBatchSpecWorkspaceFields | PreviewVisibleBatchSpecWorkspaceFields>
    /** Error */
    error?: string
    /** Whether or not the items presented in the list are read-only. */
    isReadOnly?: boolean
    /** Whether using cached results is disabled. */
    cacheDisabled?: boolean
}

export const WorkspacesPreviewList: React.FunctionComponent<React.PropsWithChildren<WorkspacesPreviewListProps>> = ({
    isStale,
    excludeRepo,
    showCached,
    cached,
    workspacesConnection: { connection, hasNextPage, fetchMore },
    error,
    cacheDisabled,
    isReadOnly,
}) => {
    const connectionOrCached = showCached && cached ? cached : connection

    return (
        <ConnectionContainer className="w-100">
            {error && <ConnectionError errors={[error]} />}
            <ConnectionList className="list-group list-group-flush w-100" aria-label="workspace results found">
                {connectionOrCached?.nodes?.map(node => (
                    <WorkspacesPreviewListItem
                        key={node.id}
                        workspace={node}
                        cacheDisabled={cacheDisabled}
                        isStale={isStale}
                        exclude={excludeRepo}
                        isReadOnly={isReadOnly}
                    />
                ))}
            </ConnectionList>
            {connectionOrCached && (
                <SummaryContainer centered={true}>
                    <ConnectionSummary
                        centered={true}
                        noSummaryIfAllNodesVisible={true}
                        connection={connectionOrCached}
                        noun="workspace"
                        pluralNoun="workspaces"
                        hasNextPage={hasNextPage}
                        emptyElement={<span className="text-muted">No workspaces found</span>}
                    />
                    {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
