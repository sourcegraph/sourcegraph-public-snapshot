import { Facet } from '@codemirror/state'

import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

export interface CodeGraphData {
    provenance: string
    commit: string
    toolInfo: {
        name: string | null
        version: string | null
    } | null
    // The raw occurrences as returned by the API
    occurrences: Occurrence[]
    // The same as occurrences, but flattened so there are no overlapping
    // ranges.
    nonOverlappingOccurrences: Occurrence[]
}

export const codeGraphData = Facet.define<CodeGraphData[], CodeGraphData[]>({
    static: true,
    combine: values => values[0] ?? [],
})
