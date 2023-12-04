import { PreciseIndexState } from '@sourcegraph/shared/src/graphql-operations'

export const INDEX_COMPLETED_STATES = new Set<PreciseIndexState>([PreciseIndexState.COMPLETED])
export const INDEX_FAILURE_STATES = new Set<PreciseIndexState>([
    PreciseIndexState.INDEXING_ERRORED,
    PreciseIndexState.PROCESSING_ERRORED,
])
export const INDEX_TERMINAL_STATES = new Set<PreciseIndexState>([...INDEX_COMPLETED_STATES, ...INDEX_FAILURE_STATES])
