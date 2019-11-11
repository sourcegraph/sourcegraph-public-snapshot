/**
 * Return the value of the given key from the given map. If the key does not
 * exist in the map, an exception is thrown with the given error text.
 *
 * @param map The map to query.
 * @param key The key to search for.
 * @param elementType The type of element (used for exception message).
 */
export function mustGet<K, V>(map: Map<K, V>, key: K, elementType: string): V {
    const value = map.get(key)
    if (value !== undefined) {
        return value
    }

    throw new Error(`Unknown ${elementType} '${key}'.`)
}

/**
 * Return the value of the given key from one of the given maps. The first
 * non-undefined value to be found is returned. If the key does not exist in
 * either map, an exception is thrown with the given error text.
 *
 * @param map1 The first map to query.
 * @param map2 The second map to query.
 * @param key The key to search for.
 * @param elementType The type of element (used for exception message).
 */
export function mustGetFromEither<K1, V1, K2, V2>(
    map1: Map<K1, V1>,
    map2: Map<K2, V2>,
    key: K1 & K2,
    elementType: string
): V1 | V2 {
    for (const map of [map1, map2]) {
        const value = map.get(key)
        if (value !== undefined) {
            return value
        }
    }

    throw new Error(`Unknown ${elementType} '${key}'.`)
}
