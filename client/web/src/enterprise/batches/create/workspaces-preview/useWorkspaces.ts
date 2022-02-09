import { dataOrThrowErrors } from '@sourcegraph/http-client'
import {
    useConnection,
    UseConnectionResult,
} from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'

import {
    Scalars,
    PreviewBatchSpecWorkspaceFields,
    BatchSpecWorkspacesPreviewResult,
    BatchSpecWorkspacesPreviewVariables,
} from '../../../../graphql-operations'
import { WORKSPACES } from '../backend'

export const WORKSPACES_PER_PAGE_COUNT = 100

export interface WorkspacePreviewFilters {
    search: string | null
}

/**
 * Custom hook that wraps `useConnection` to resolve the workspaces for the batch spec
 * with the ID and filters provided.
 *
 * @param batchSpecID The id of the batch spec to query
 * @param filters Any filters to apply to the workspaces connection preview
 */
export const useWorkspaces = (
    batchSpecID: Scalars['ID'],
    filters?: WorkspacePreviewFilters
): UseConnectionResult<PreviewBatchSpecWorkspaceFields> =>
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
