import { BloomFilter } from 'bloomfilter'

// These parameters give us a 1 in 1.38x10^9 false positive rate if we assume
// that the number of unique URIs referrable by an external package is of the
// order of 10k (....but I have no idea if that is a reasonable estimate....).
//
// See the following link for a bloom calculator: https://hur.st/bloomfilter
const BLOOM_FILTER_BITS = 64 * 1024
const BLOOM_FILTER_NUM_HASH_FUNCTIONS = 16

export function testFilter(filter: string, uri: string): boolean {
    // TOOD(efritz) - decode smarter
    return new BloomFilter(JSON.parse(filter), BLOOM_FILTER_NUM_HASH_FUNCTIONS).test(uri)
}

export function createFilter(uris: string[]): string {
    // TODO(efritz) - encode smarter
    const filter = new BloomFilter(BLOOM_FILTER_BITS, BLOOM_FILTER_NUM_HASH_FUNCTIONS)
    uris.forEach(uri => filter.add(uri))
    return JSON.stringify([].slice.call(filter.buckets))
}
