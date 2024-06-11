// TODO: document that this is the new way to do things

import { Facet } from '@codemirror/state'

import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

export interface CodeGraphData {
    // TODO: add other relevant information here?
    provenance: string
    occurrences: Occurrence[]
}

export const codeGraphData = Facet.define<CodeGraphData[], CodeGraphData[]>({
    static: true,
    // TODO: generate an index for efficient lookups
    combine: values => values[0],
})
