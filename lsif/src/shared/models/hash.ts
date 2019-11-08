import * as dumpModels from './dump'

/**
 * Hash a string or numeric identifier into the range `[0, maxIndex)`. The
 * hash algorithm here is similar to the one used in Java's String.hashCode.
 *
 * @param id The identifier to hash.
 * @param maxIndex The maximum of the range.
 */
export function hashKey(id: dumpModels.DefinitionReferenceResultId, maxIndex: number): number {
    const s = `${id}`

    let hash = 0
    for (let i = 0; i < s.length; i++) {
        const chr = s.charCodeAt(i)
        hash = (hash << 5) - hash + chr
        hash |= 0
    }

    // Hash value may be negative - must unset sign bit before modulus
    return Math.abs(hash) % maxIndex
}
