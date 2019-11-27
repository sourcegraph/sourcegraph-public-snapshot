/**
 * Parse a base64-encoded JSON payload into the expected type.
 *
 * @param cursorRaw The raw cursor.
 */
export function parseCursor<T>(cursorRaw: string | undefined): T | undefined {
    if (cursorRaw === undefined) {
        return undefined
    }

    try {
        // False positive https://github.com/typescript-eslint/typescript-eslint/issues/1269
        // eslint-disable-next-line @typescript-eslint/return-await
        return JSON.parse(Buffer.from(cursorRaw, 'base64').toString())
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

    return Buffer.from(JSON.stringify(cursor)).toString('base64')
}
