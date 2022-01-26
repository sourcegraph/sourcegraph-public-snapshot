import React from 'react'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import {
    useConnection,
    UseConnectionResult,
} from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import {
    Scalars,
    PreviewBatchSpecWorkspaceFields,
    BatchSpecWorkspacesPreviewResult,
    BatchSpecWorkspacesPreviewVariables,
} from '../../../../graphql-operations'
import { WORKSPACES } from '../backend'

import { PreviewLoadingSpinner } from './PreviewLoadingSpinner'
import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

interface WorkspacesPreviewListProps {
    batchSpecID: Scalars['ID']
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
}

const WORKSPACES_PER_PAGE_COUNT = 100

export const WorkspacesPreviewList: React.FunctionComponent<WorkspacesPreviewListProps> = ({
    batchSpecID,
    isStale,
    excludeRepo,
}) => {
    const { connection, error, loading, hasNextPage, fetchMore } = useWorkspaces(batchSpecID)

    if (loading) {
        return <PreviewLoadingSpinner className="my-4" />
    }

    return (
        <ConnectionContainer className="w-100">
            {error && <ConnectionError errors={[error.message]} />}
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
            {loading && <ConnectionLoading />}
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

const useWorkspaces = (batchSpecID: Scalars['ID']): UseConnectionResult<PreviewBatchSpecWorkspaceFields> =>
    useConnection<
        BatchSpecWorkspacesPreviewResult,
        BatchSpecWorkspacesPreviewVariables,
        PreviewBatchSpecWorkspaceFields
    >({
        query: WORKSPACES,
        variables: {
            batchSpec: batchSpecID,
            after: null,
            first: WORKSPACES_PER_PAGE_COUNT,
        },
        options: {
            useURL: false,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.node) {
                throw new Error(`Batch spec with ID ${batchSpecID} does not exist`)
            }
            if (data.node.__typename !== 'BatchSpec') {
                throw new Error(`The given ID is a ${data.node.__typename as string}, not a BatchSpec`)
            }
            if (!data.node.workspaceResolution) {
                throw new Error(`No workspace resolution found for batch spec with ID ${batchSpecID}`)
            }
            return data.node.workspaceResolution.workspaces
        },
    })
