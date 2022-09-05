import { BatchSpecWorkspaceState } from '../../../../graphql-operations'

export function isValidBatchSpecWorkspaceState(state: string): state is BatchSpecWorkspaceState {
    return Object.values(BatchSpecWorkspaceState).some(value => value === state)
}
