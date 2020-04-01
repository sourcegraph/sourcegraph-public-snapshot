import * as lsp from 'vscode-languageserver-protocol'
import { OrderedSet } from '../../shared/datastructures/orderedset'

export interface InternalLocation {
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
