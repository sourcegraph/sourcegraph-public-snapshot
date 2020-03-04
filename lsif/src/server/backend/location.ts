import * as lsp from 'vscode-languageserver-protocol'
import * as pgModels from '../../shared/models/pg'

/** A location with the dump that contains it. */
export interface InternalLocation {
    dump: pgModels.LsifDump
    path: string
    range: lsp.Range
}

/**
 * Returns a new list of locations with duplicates removes. Will preserve the order
 * of the original list. The first instance of a location to be seen will remain.
 *
 * This method is used in place of lodash.uniqWith in some places due to efficiency
 * concerns. Some dumps in the query path may return tens or hundreds of thousands
 * of reference records on a moniker search. This method makes a single linear pass
 * over the elements of the array, where uniqWith is implemented with a quadratic
 * nested loop.
 *
 * @param locations The locations to deduplicate.
 */
export function deduplicateLocations(locations: InternalLocation[]): InternalLocation[] {
    const seen = new Set<string>()

    return locations.filter(location => {
        const key = makeKey(location)
        if (seen.has(key)) {
            return false
        }

        seen.add(key)
        return true
    })
}

/** Makes a unique string representation of this location. */
function makeKey(location: InternalLocation): string {
    return [
        location.dump.id,
        location.path,
        location.range.start.line,
        location.range.start.character,
        location.range.end.line,
        location.range.end.character,
    ].join(':')
}
