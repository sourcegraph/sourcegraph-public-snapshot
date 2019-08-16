import { gzip, gunzip } from 'mz/zlib'
import * as crypto from 'crypto'
import { BloomFilter } from 'bloomfilter'

// These parameters give us a 1 in 1.38x10^9 false positive rate if we assume
// that the number of unique URIs referrable by an external package is of the
// order of 10k (....but I have no idea if that is a reasonable estimate....).
//
// See the following link for a bloom calculator: https://hur.st/bloomfilter
const BLOOM_FILTER_BITS = 64 * 1024
const BLOOM_FILTER_NUM_HASH_FUNCTIONS = 16

/**
 * Create a bloom filter containing the given values and return it as a base64
 * encoded gzipped string.
 *
 * @param uris The values to add to the bloom filter.
 */
export function createFilter(uris: string[]): Promise<string> {
    const filter = new BloomFilter(BLOOM_FILTER_BITS, BLOOM_FILTER_NUM_HASH_FUNCTIONS)
    uris.forEach(uri => filter.add(uri))
    const buckets = [].slice.call(filter.buckets)
    return encodeJSON(buckets)
}

/**
 * Decode `filter` as created by `createFilter` and determine if `uri` is a
 * possible element. This may return a false positive (returning true if the
 * element is not actually a member), but will not return false negatives.
 *
 * @param filter The encoded filter.
 * @param uri The uri to test membership.
 */
export async function testFilter(filter: string, uri: string): Promise<boolean> {
    return new BloomFilter(await decodeJSON(filter), BLOOM_FILTER_NUM_HASH_FUNCTIONS).test(uri)
}

/**
 * Return the output of both `hashJSON` and `encodeJSON` on `value`.
 *
 * @param value The value to hash and encode.
 */
export async function hashAndEncodeJSON<T>(value: T): Promise<{ hash: string; encoded: string }> {
    const jsonified = jsonify(value)

    return {
        hash: hash(jsonified),
        encoded: await encode(jsonified),
    }
}

/**
 * Return the hash of the JSON representation of `value`.
 *
 * @param value The value to hash.
 */
export function hashJSON<T>(value: T): string {
    return hash(jsonify(value))
}

/**
 * Return the hash of `value.
 *
 * @param value The value to hash.
 */
export function hash(value: string): string {
    const hash = crypto.createHash('md5')
    hash.update(value)
    return hash.digest('base64')
}

/**
 * Return the base64-encoded gzipped JSON representation of `value`.
 *
 * @param value The value to encode.
 */
export function encodeJSON<T>(value: T): Promise<string> {
    return encode(jsonify(value))
}

/**
 * Return the base64-encoded gzipped representation of `value`.
 *
 * @param value The value to encode.
 */
export async function encode(value: string): Promise<string> {
    return (await gzip(Buffer.from(value))).toString('base64')
}

/**
 * Reverse the operation of `encodeJSON`.
 *
 * @param value The value to decode.
 */
export async function decodeJSON<T>(value: string): Promise<T> {
    return JSON.parse(await decode(value))
}

/**
 * Reverse the operation of `encode`.
 *
 * @param value The value to decode.
 */
export async function decode(value: string): Promise<string> {
    return (await gunzip(Buffer.from(value, 'base64'))).toString()
}

/**
 * Return the JSON representation of `value`.
 *
 * @param value The value to jsonify.
 */
function jsonify<T>(value: T): string {
    return JSON.stringify(value, undefined, 0)
}
