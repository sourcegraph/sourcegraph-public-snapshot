import { dataOrThrowErrors } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../../../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    Scalars,
    BatchSpecWorkspacesPreviewResult,
    BatchSpecWorkspacesPreviewVariables,
    PreviewHiddenBatchSpecWorkspaceFields,
    PreviewVisibleBatchSpecWorkspaceFields,
} from '../../../../../graphql-operations'
import { WORKSPACES } from '../../../create/backend'

export const WORKSPACES_PER_PAGE_COUNT = 100

/** Any filters to apply to the workspaces preview connection. */
export interface WorkspacePreviewFilters {
    search: string | null
}

/**
 * Custom hook that wraps `useShowMorePagination` to resolve the workspaces for the batch spec
 * with the ID and filters provided.
 *
 * @param batchSpecID The id of the batch spec to query.
 * @param filters Any filters to apply to the workspaces preview connection.
 */
export const useWorkspaces = (
    batchSpecID: Scalars['ID'],
    filters?: WorkspacePreviewFilters
): UseShowMorePaginationResult<
    BatchSpecWorkspacesPreviewResult,
    PreviewHiddenBatchSpecWorkspaceFields | PreviewVisibleBatchSpecWorkspaceFields
> =>
    useShowMorePagination<
        BatchSpecWorkspacesPreviewResult,
        BatchSpecWorkspacesPreviewVariables,
        PreviewHiddenBatchSpecWorkspaceFields | PreviewVisibleBatchSpecWorkspaceFields
    >({
        query: WORKSPACES,
        variables: {
            batchSpec: batchSpecID,
            after: null,
            first: WORKSPACES_PER_PAGE_COUNT,
            search: filters?.search ?? null,
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
                return { nodes: [] }
            }

            return data.node.workspaceResolution.workspaces
        },
    })
