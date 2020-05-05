import { readEnvInt } from '../settings'

// These parameters give us a 1 in 1.38x10^9 false positive rate if we assume
// that the number of unique URIs referrable by an external package is of the
// order of 10k (....but I have no idea if that is a reasonable estimate....).
//
// See the following link for a bloom calculator: https://hur.st/bloomfilter

/** The number of bits allocated for new bloom filters. */
export const BLOOM_FILTER_BITS = readEnvInt('BLOOM_FILTER_BITS', 64 * 1024)

/** The number of hash functions to use to determine if a value is a member of the filter. */
export const BLOOM_FILTER_NUM_HASH_FUNCTIONS = readEnvInt('BLOOM_FILTER_NUM_HASH_FUNCTIONS', 16)
