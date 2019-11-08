import { BloomFilter } from 'bloomfilter'
import { gunzipJSON, gzipJSON } from '../encoding/json'
import { readEnvInt } from '../settings'

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
