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
    /** Cached */
    showCached: boolean
    cached?: PreviewBatchSpecWorkspaceFields[]
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
    if (showCached) {
        return (
            <ConnectionContainer className="w-100">
                {error && <ConnectionError errors={[error]} />}
                <ConnectionList className="list-group list-group-flush w-100">
                    {cached?.map((node, index) => (
                        <WorkspacesPreviewListItem
                            key={`${node.repository.id}-${node.branch.id}`}
                            item={node}
                            isStale={isStale}
                            exclude={excludeRepo}
                            variant={index % 2 === 0 ? 'light' : 'dark'}
                        />
                    ))}
                </ConnectionList>
            </ConnectionContainer>
        )
    }

    return (
        <ConnectionContainer className="w-100">
            {error && <ConnectionError errors={[error]} />}
            <ConnectionList className="list-group list-group-flush w-100">
                {connection?.nodes?.map((node, index) => (
                    <WorkspacesPreviewListItem
                        key={`${node.repository.id}-${node.branch.id}`}
                        item={node}
                        isStale={isStale}
                        exclude={excludeRepo}
                        variant={index % 2 === 0 ? 'light' : 'dark'}
                    />
                ))}
            </ConnectionList>
            {connection && (
                <SummaryContainer centered={true}>
                    <ConnectionSummary
                        noSummaryIfAllNodesVisible={true}
                        first={WORKSPACES_PER_PAGE_COUNT}
                        connection={connection}
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
