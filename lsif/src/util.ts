import { Id } from 'lsif-protocol'

/**
 * Reads an integer from an environment variable or defaults to the given value.
 *
 * @param key The environment variable name.
 * @param defaultValue The default value.
 */
export function readEnvInt(key: string, defaultValue: number): number {
    return (process.env[key] && parseInt(process.env[key] || '', 10)) || defaultValue
}

/**
 * Determine if an exception value has the given error code.
 *
 * @param e The exception value.
 * @param expectedCode The expected error code.
 */
export function hasErrorCode(e: any, expectedCode: string): boolean {
    return e && e.code === expectedCode
}

/**
 * Return the value of the given key in one of the given maps. The first value
 * to exist is returned. If the key does not exist in any map, an exception is
 * thrown.
 *
 * @param key The key to search for.
 * @param name The type of element (used for exception message).
 * @param maps The set of maps to query.
 */
export function assertDefined<K, V>(key: K, name: string, ...maps: Map<K, V>[]): V {
    for (const map of maps) {
        const value = map.get(key)
        if (value !== undefined) {
            return value
        }
    }

    throw new Error(`Unknown ${name} '${key}'.`)
}

/**
 * Return the value of `id`, or throw an exception if it is undefined.
 *
 * @param id The identifier.
 */
export function assertId(id: Id | undefined): Id {
    if (id !== undefined) {
        return id
    }

    throw new Error('id is undefined')
}

/**
 * Hash a string or numeric identifier into the range [0, `maxIndex`). The
 * hash algorithm here is similar to the one used in Java's String.hashCode.
 *
 * @param id The identifier to hash.
 * @param maxIndex The maximum of the range.
 */
export function hashKey(id: Id, maxIndex: number): number {
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
