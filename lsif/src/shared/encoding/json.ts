import { gunzip, gzip } from 'mz/zlib'

/**
 * Return the gzipped JSON representation of `value`.
 *
 * @param value The value to encode.
 */
export function gzipJSON<T>(value: T): Promise<Buffer> {
    return gzip(Buffer.from(dumpJSON(value)))
}

/**
 * Reverse the operation of `gzipJSON`.
 *
 * @param value The value to decode.
 */
export async function gunzipJSON<T>(value: Buffer): Promise<T> {
    return parseJSON((await gunzip(value)).toString())
}

/**
 * Return the JSON representation of `value`. This has special logic to
 * convert ES6 map and set structures into a JSON-representable value.
 * This method, along with `parseJSON` should be used over the raw methods
 * if the payload may contain maps.
 *
 * @param value The value to jsonify.
 */
function dumpJSON<T>(value: T): string {
    return JSON.stringify(value, (_, oldValue) => {
        if (oldValue instanceof Map) {
            return {
                type: 'map',
                value: [...oldValue],
            }
        }

        if (oldValue instanceof Set) {
            return {
                type: 'set',
                value: [...oldValue],
            }
        }

        return oldValue
    })
}

/**
 * Parse the JSON representation of `value`. This has special logic to
 * unmarshal map and set objects as encoded by `dumpJSON`.
 *
 * @param value The value to unmarshal.
 */
function parseJSON<T>(value: string): T {
    return JSON.parse(value, (_, oldValue) => {
        if (typeof oldValue === 'object' && oldValue !== null) {
            if (oldValue.type === 'map') {
                return new Map(oldValue.value)
            }

            if (oldValue.type === 'set') {
                return new Set(oldValue.value)
            }
        }

        return oldValue
    })
}
