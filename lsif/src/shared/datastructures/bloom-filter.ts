import { BloomFilter } from 'bloomfilter'
import { gunzipJSON, gzipJSON } from '../encoding/json'
import * as settings from './settings'

/** A type that describes a the encoded version of a bloom filter. */
export type EncodedBloomFilter = Buffer

/**
 * Create a bloom filter containing the given values and return an encoded verion.
 *
 * @param values The values to add to the bloom filter.
 */
export function createFilter(values: string[]): Promise<EncodedBloomFilter> {
    const filter = new BloomFilter(settings.BLOOM_FILTER_BITS, settings.BLOOM_FILTER_NUM_HASH_FUNCTIONS)
    for (const value of values) {
        filter.add(value)
    }

    // Need to shed the type of the array
    const buckets = Array.from(filter.buckets)

    // Store the number of hash functions used to create this as it may change after
    // this value is serialized. We don't want to test with more hash functions than
    // it was created with, otherwise we'll get false negatives.
    return gzipJSON({ numHashFunctions: settings.BLOOM_FILTER_NUM_HASH_FUNCTIONS, buckets })
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
