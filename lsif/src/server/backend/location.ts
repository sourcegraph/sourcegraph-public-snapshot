import * as lsp from 'vscode-languageserver-protocol'
import * as pgModels from '../../shared/models/pg'
import { OrderedSet } from '../../shared/datastructures/orderedset'

export interface InternalLocation {
    /** The identifier of the dump that contains the location. */
    dumpId: pgModels.DumpId
    /** The path relative to the dump root. */
    path: string
    range: lsp.Range
}

export interface ResolvedInternalLocation {
    /** The dump that contains the location. */
    dump: pgModels.LsifDump
    /** The path relative to the dump root. */
    path: string
    range: lsp.Range
}

/** A duplicate-free list of locations ordered by time of insertion. */
export class OrderedLocationSet extends OrderedSet<InternalLocation> {
    /**
     * Create a new ordered locations set.
     *
     * @param values A set of values used to seed the set.
     */
    constructor(values?: InternalLocation[]) {
        super(
            (value: InternalLocation): string =>
                [
                    value.dumpId,
                    value.path,
                    value.range.start.line,
                    value.range.start.character,
                    value.range.end.line,
                    value.range.end.character,
                ].join(':'),
            values
        )
    }
}
