import { BatchSpecWorkspaceResolutionState, WorkspaceResolutionStatusResult } from '../../../graphql-operations'

export const mockWorkspaceResolutionStatus = (
    status: BatchSpecWorkspaceResolutionState,
    error?: string
): WorkspaceResolutionStatusResult => ({
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: {
            __typename: 'BatchSpecWorkspaceResolution',
            state: status,
            failureMessage: error || null,
        },
    },
})
