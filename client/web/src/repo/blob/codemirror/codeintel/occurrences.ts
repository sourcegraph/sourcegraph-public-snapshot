import { Facet } from '@codemirror/state'

import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

export interface CodeGraphData {
    provenance: string
    commit: string
    toolInfo: {
        name: string | null
        version: string | null
    } | null
    // The raw occurrences as returned by the API. Guaranteed to be sorted.
    occurrences: Occurrence[]
    // The same as occurrences, but flattened so there are no overlapping
    // ranges. Guaranteed to be sorted.
    nonOverlappingOccurrences: Occurrence[]
}

// A facet that contains the precise code graph data from the occurrences API.
// It just retains the most recent contribution. At some point, we should
// probably extend this to be able to accept contributions from multiple
// sources.
export const codeGraphData = Facet.define<CodeGraphData[], CodeGraphData[]>({
    combine: values => values[0] ?? [],
})
