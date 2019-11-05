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
 * A type that describes a the encoded version of a bloom filter.
 */
export type EncodedBloomFilter = Buffer

/**
 * Create a bloom filter containing the given values and return an encoded verion.
 *
 * @param values The values to add to the bloom filter.
 */
export function createFilter(values: string[]): Promise<EncodedBloomFilter> {
    const filter = new BloomFilter(BLOOM_FILTER_BITS, BLOOM_FILTER_NUM_HASH_FUNCTIONS)
    for (const value of values) {
        filter.add(value)
    }

    // Need to shed the type of the array
    const buckets = Array.from(filter.buckets)

    // Store the number of hash functions used to create this as it may change after
    // this value is serialized. We don't want to test with more hash functions than
    // it was created with, otherwise we'll get false negatives.
    return gzipJSON({ numHashFunctions: BLOOM_FILTER_NUM_HASH_FUNCTIONS, buckets })
}

/**
 * Decode `filter` as created by `createFilter` and determine if `value` is a
 * possible element. This may return a false positive (returning true if the
 * element is not actually a member), but will not return false negatives.
 *
 * @param filter The encoded filter.
 * @param value The value to test membership.
 */
export async function testFilter(filter: EncodedBloomFilter, value: string): Promise<boolean> {
    const { numHashFunctions, buckets } = await gunzipJSON(filter)
    return new BloomFilter(buckets, numHashFunctions).test(value)
}

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
