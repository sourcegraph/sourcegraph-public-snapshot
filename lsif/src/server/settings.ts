import { readEnvInt } from '../shared/settings'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
export const HTTP_PORT = readEnvInt('HTTP_PORT', 3186)

/**
 * The host and port running the redis instance containing work queues.
 *
 * Set addresses. Prefer in this order:
 *   - Specific envvar REDIS_STORE_ENDPOINT
 *   - Fallback envvar REDIS_ENDPOINT
 *   - redis-store:6379
 *
 *  Additionally keep this logic in sync with pkg/redispool/redispool.go and cmd/server/redis.go
 */
export const REDIS_ENDPOINT = process.env.REDIS_STORE_ENDPOINT || process.env.REDIS_ENDPOINT || 'redis-store:6379'

/**
 * Where on the file system to store LSIF files.
 */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * The interval (in seconds) to invoke the cleanOldUploads task.
 */
export const CLEAN_OLD_UPLOADS_INTERVAL = readEnvInt('CLEAN_OLD_UPLOADS_INTERVAL', 60 * 60 * 8)

/**
 * The default number of remote dumps to open when performing a global find-reference operation.
 */
export const DEFAULT_REFERENCES_NUM_REMOTE_DUMPS = readEnvInt('DEFAULT_REFERENCES_NUM_REMOTE_DUMPS', 10)

/**
 * The interval (in seconds) to invoke the cleanFailedUploads task.
 */
export const CLEAN_FAILED_UPLOADS_INTERVAL = readEnvInt('CLEAN_FAILED_UPLOADS_INTERVAL', 60 * 60 * 8)

/**
 * The interval (in seconds) to invoke the updateQueueSizeGaugeInterval task.
 */
export const UPDATE_QUEUE_SIZE_GAUGE_INTERVAL = readEnvInt('UPDATE_QUEUE_SIZE_GAUGE_INTERVAL', 5)

/**
 * The interval (in seconds) to run the resetStalledUploads task.
 */
export const RESET_STALLED_UPLOADS_INTERVAL = readEnvInt('RESET_STALLED_UPLOADS_INTERVAL', 60)

/**
 * The default page size for the job endpoints.
 */
export const DEFAULT_JOB_PAGE_SIZE = readEnvInt('DEFAULT_JOB_PAGE_SIZE', 50)

/**
 * The default page size for the upload endpoints.
 */
export const DEFAULT_UPLOAD_PAGE_SIZE = readEnvInt('DEFAULT_UPLOAD_PAGE_SIZE', 50)

/**
 * The maximum number of jobs to search in one call to the search-jobs.lua script.
 */
export const MAX_JOB_SEARCH = readEnvInt('MAX_JOB_SEARCH', 10000)

/**
 * The default page size for the dumps endpoint.
 */
export const DEFAULT_DUMP_PAGE_SIZE = readEnvInt('DEFAULT_DUMP_PAGE_SIZE', 50)

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
export const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 100)

/**
 * The maximum number of documents that can be held in memory at once.
 */
export const DOCUMENT_CACHE_CAPACITY = readEnvInt('DOCUMENT_CACHE_CAPACITY', 1024 * 1024 * 1024)

/**
 * The maximum number of result chunks that can be held in memory at once.
 */
export const RESULT_CHUNK_CACHE_CAPACITY = readEnvInt('RESULT_CHUNK_CACHE_CAPACITY', 1024 * 1024 * 1024)

/**
 * The maximum age (in seconds) that an upload (completed or queued) will remain in Postgres.
 */
export const UPLOAD_MAX_AGE = readEnvInt('JOB_UPLOAD_AGE', 60 * 60 * 24 * 7)

/**
 * The maximum age (in seconds) that the files for an unprocessed upload can remain on disk.
 */
export const FAILED_UPLOAD_MAX_AGE = readEnvInt('FAILED_UPLOAD_MAX_AGE', 24 * 60 * 60)

/**
 * The maximum age (in seconds) that the an upload can be unlocked and in the `processing` state.
 */
export const STALLED_UPLOAD_MAX_AGE = readEnvInt('STALLED_UPLOAD_MAX_AGE', 5)
