/**
 * Whether or not to log a message when the HTTP server is ready and listening.
 */
export const LOG_READY = process.env.DEPLOY_TYPE === 'dev'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
export const HTTP_PORT = readEnvInt('LSIF_HTTP_PORT', 3186)

/**
 * The host running the redis instance containing work queues. Defaults to localhost.
 */
export const REDIS_HOST = process.env.LSIF_REDIS_HOST || 'localhost'

/**
 * The port of the redis instance containing work queues. Defaults to 6379.
 */
export const REDIS_PORT = readEnvInt('LSIF_REDIS_PORT', 6379)

/**
 * Where on the file system to store LSIF files.
 */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

//
// Caches

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
export const CONNECTION_CACHE_SIZE = readEnvInt('CONNECTION_CACHE_SIZE', 1000)

/**
 * The maximum number of documents that can be held in memory at once.
 */
export const DOCUMENT_CACHE_SIZE = readEnvInt('DOCUMENT_CACHE_SIZE', 1000)

//
// Bloom Filter

// These parameters give us a 1 in 1.38x10^9 false positive rate if we assume
// that the number of unique URIs referrable by an external package is of the
// order of 10k (....but I have no idea if that is a reasonable estimate....).
//
// See the following link for a bloom calculator: https://hur.st/bloomfilter

/**
 * The number of bits allocated for new bloom filters.
 */
export const BLOOM_FILTER_BITS = readEnvInt('BLOOM_FILTER_BITS', 64 * 1024)

/**
 * The number of hash functions to use to determine if a value is a member of the filter.
 */
export const BLOOM_FILTER_NUM_HASH_FUNCTIONS = readEnvInt('BLOOM_FILTER_NUM_HASH_FUNCTIONS', 16)

//
// Helpers

/**
 * Reads an integer from an environment variable or defaults to the given value.
 *
 * @param key The environment variable name.
 * @param defaultValue The default value.
 */
function readEnvInt(key: string, defaultValue: number): number {
    return (process.env[key] && parseInt(process.env[key] || '', 10)) || defaultValue
}
