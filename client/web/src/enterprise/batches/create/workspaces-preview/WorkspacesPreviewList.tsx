import React from 'react'

import { UseConnectionResult } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import { Connection } from '../../../../components/FilteredConnection'
import { PreviewBatchSpecWorkspaceFields } from '../../../../graphql-operations'

import { WORKSPACES_PER_PAGE_COUNT } from './useWorkspaces'
import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

interface WorkspacesPreviewListProps {
    workspacesConnection: UseConnectionResult<PreviewBatchSpecWorkspaceFields>
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
    cached?: Connection<PreviewBatchSpecWorkspaceFields>
    /** Error */
    error?: string
}

export const WorkspacesPreviewList: React.FunctionComponent<WorkspacesPreviewListProps> = ({
    isStale,
    excludeRepo,
    showCached,
    cached,
    workspacesConnection: { connection, hasNextPage, fetchMore },
    error,
}) => {
    const connectionOrCached = showCached && cached ? cached : connection

    return (
        <ConnectionContainer className="w-100">
            {error && <ConnectionError errors={[error]} />}
            <ConnectionList className="list-group list-group-flush w-100">
                {connectionOrCached?.nodes?.map((node, index) => (
                    <WorkspacesPreviewListItem
                        key={`${node.repository.id}-${node.branch.id}`}
                        item={node}
                        isStale={isStale}
                        exclude={excludeRepo}
                        variant={index % 2 === 0 ? 'light' : 'dark'}
                    />
                ))}
            </ConnectionList>
            {connectionOrCached && (
                <SummaryContainer centered={true}>
                    <ConnectionSummary
                        noSummaryIfAllNodesVisible={true}
                        first={WORKSPACES_PER_PAGE_COUNT}
                        connection={connectionOrCached}
                        noun="workspace"
                        pluralNoun="workspaces"
                        hasNextPage={hasNextPage}
                        emptyElement={<span className="text-muted">No workspaces found</span>}
                    />
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
