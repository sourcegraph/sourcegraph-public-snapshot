// TODO: document that this is the new way to do things

import { Facet } from '@codemirror/state'

import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

export interface CodeGraphData {
    provenance: string
    toolInfo?: {
        name?: string
        version?: string
    }
    // Guaranteed to be sorted by range
    occurrences: Occurrence[]
}

export const codeGraphData = Facet.define<CodeGraphData[], CodeGraphData[]>({
    static: true,
    combine: values => values[0] ?? [],
})
