/**
 * Parse a base64-encoded JSON payload into the expected type.
 *
 * @param cursorRaw The raw cursor.
 */
export function parseCursor<T>(cursorRaw: any): T | undefined {
    if (cursorRaw === undefined) {
        return undefined
    }

    try {
        return JSON.parse(new Buffer(cursorRaw, 'base64').toString('ascii'))
    } catch {
        throw Object.assign(new Error(`Malformed cursor supplied ${cursorRaw}`), { status: 400 })
    }
}

/**
 * Encode an arbitrary pagination cursor value into a string.
 *
 * @param cursor The cursor value.
 */
export function encodeCursor<T>(cursor: T | undefined): string | undefined {
    if (!cursor) {
        return undefined
    }

    return new Buffer(JSON.stringify(cursor)).toString('base64')
}
