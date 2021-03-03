/**
 * Returns a hash code value for a string.
 *
 * The hash algorithm here is similar to the one used in Java's String.hashCode
 * and in lsifstore.
 */
export function hashCode(string: string, maxIndex: number): number {
    let hash = 0

    for (let index = 0; index < string.length; index++) {
        const char = string.charCodeAt(index)
        hash = (hash << 5) - hash + char
        hash |= 0 // Convert to 32bit integer
    }

    if (hash < 0) {
        hash = -hash
    }

    return hash % maxIndex
}
