import * as lsp from 'vscode-languageserver-protocol'
import { OrderedSet } from '../../shared/datastructures/orderedset'

export interface Location {
    path: string
    range: lsp.Range
}

/** A duplicate-free list of locations ordered by time of insertion. */
export class OrderedLocationSet extends OrderedSet<Location> {
    /**
     * Create a new ordered locations set.
     *
     * @param values A set of values used to seed the set.
     * @param trusted Whether the given values are already deduplicated.
     */
    constructor(locations?: Location[], trusted = false) {
        super(
            (location: Location): string =>
                [
                    location.path,
                    location.range.start.line,
                    location.range.start.character,
                    location.range.end.line,
                    location.range.end.character,
                ].join(':'),
            locations,
            trusted
        )
    }
}
