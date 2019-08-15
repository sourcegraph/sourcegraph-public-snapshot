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

export async function testFilter(filter: string, uri: string): Promise<boolean> {
    return new BloomFilter(await decodeJSON(filter), BLOOM_FILTER_NUM_HASH_FUNCTIONS).test(uri)
}

export function createFilter(uris: string[]): Promise<string> {
    const filter = new BloomFilter(BLOOM_FILTER_BITS, BLOOM_FILTER_NUM_HASH_FUNCTIONS)
    uris.forEach(uri => filter.add(uri))
    const buckets = [].slice.call(filter.buckets)
    return encodeJSON(buckets)
}

export async function hashAndEncodeJSON<T>(value: T): Promise<{ hash: string; encoded: string }> {
    const jsonified = jsonify(value)

    return {
        hash: hash(jsonified),
        encoded: await encode(jsonified),
    }
}

export function hashJSON<T>(value: T): string {
    return hash(jsonify(value))
}

export function hash(value: string): string {
    const hash = crypto.createHash('md5')
    hash.update(value)
    return hash.digest('base64')
}

export function encodeJSON<T>(value: T): Promise<string> {
    return encode(jsonify(value))
}

export async function encode(value: string): Promise<string> {
    return (await gzip(Buffer.from(value))).toString('base64')
}

export async function decodeJSON<T>(value: string): Promise<T> {
    return JSON.parse(await decode(value))
}

export async function decode(value: string): Promise<string> {
    return (await gunzip(Buffer.from(value, 'base64'))).toString()
}

function jsonify<T>(value: T): string {
    return JSON.stringify(value, undefined, 0)
}
