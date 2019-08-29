import { gzip, gunzip } from 'mz/zlib'
import { BloomFilter } from 'bloomfilter'
import { readEnvInt } from './util'

// These parameters give us a 1 in 1.38x10^9 false positive rate if we assume
// that the number of unique URIs referrable by an external package is of the
// order of 10k (....but I have no idea if that is a reasonable estimate....).
//
// See the following link for a bloom calculator: https://hur.st/bloomfilter

/**
 * The number of bits allocated for new bloom filters.
 */
const BLOOM_FILTER_BITS = readEnvInt('BLOOM_FILTER_BITS', 64 * 1024)

/**
 * The number of hash functions to use to determine if a value is a member of the filter.
 */
const BLOOM_FILTER_NUM_HASH_FUNCTIONS = readEnvInt('BLOOM_FILTER_NUM_HASH_FUNCTIONS', 16)

/**
 * Create a bloom filter containing the given values and return it as a base64
 * encoded gzipped string.
 *
 * @param values The values to add to the bloom filter.
 */
export function createFilter(values: string[]): Promise<string> {
    const filter = new BloomFilter(BLOOM_FILTER_BITS, BLOOM_FILTER_NUM_HASH_FUNCTIONS)
    for (const value of values) {
        filter.add(value)
    }

    // Need to shed the type of the array
    const buckets = Array.from(filter.buckets)

    // Store the number of hash functions used to create this as it may change after
    // this value is serialized. We don't want to test with more hash functions than
    // it was created with, otherwise we'll get false negatives.
    return encodeJSON({ numHashFunctions: BLOOM_FILTER_NUM_HASH_FUNCTIONS, buckets })
}

/**
 * Decode `filter` as created by `createFilter` and determine if `value` is a
 * possible element. This may return a false positive (returning true if the
 * element is not actually a member), but will not return false negatives.
 *
 * @param filter The encoded filter.
 * @param value The value to test membership.
 */
export async function testFilter(filter: string, value: string): Promise<boolean> {
    const { numHashFunctions, buckets } = await decodeJSON(filter)
    return new BloomFilter(buckets, numHashFunctions).test(value)
}

/**
 * Return the base64-encoded gzipped JSON representation of `value`.
 *
 * @param value The value to encode.
 */
export function encodeJSON<T>(value: T): Promise<string> {
    return b64gzip(dumpJSON(value))
}

/**
 * Reverse the operation of `encodeJSON`.
 *
 * @param value The value to decode.
 */
export async function decodeJSON<T>(value: string): Promise<T> {
    return parseJSON(await unb64gzip(value))
}

/**
 * Return the base64-encoded gzipped representation of `value`.
 *
 * @param value The value to encode.
 */
async function b64gzip(value: string): Promise<string> {
    return (await gzip(Buffer.from(value))).toString('base64')
}

/**
 * Reverse the operation of `b64gzip`.
 *
 * @param value The value to decode.
 */
async function unb64gzip(value: string): Promise<string> {
    return (await gunzip(Buffer.from(value, 'base64'))).toString()
}

/**
 * Return the JSON representation of `value`. This has special logic to
 * convert an ES6 map structure into a JSON-representable value. This
 * method, along with `parseJSON` should be used over the raw methods if
 * the payload may contain maps.
 *
 * @param value The value to jsonify.
 */
function dumpJSON<T>(value: T): string {
    return JSON.stringify(value, (_, value) => {
        if (value instanceof Map) {
            return {
                type: 'map',
                value: [...value],
            }
        }

        return value
    })
}

/**
 * Parse the JSON representation of `value`. This has special logic to
 * unmarshal map objects as encoded by `dumpJSON`.
 *
 * @param value The value to unmarshal.
 */
function parseJSON<T>(value: string): T {
    return JSON.parse(value, (_, value) => {
        if (typeof value === 'object' && value !== null && value.type === 'map') {
            return new Map(value.value)
        }

        return value
    })
}
