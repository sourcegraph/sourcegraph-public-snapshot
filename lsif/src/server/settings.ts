import { readEnvInt } from '../shared/util'

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
 * The interval (in seconds) to schedule the update-tips job.
 */
export const UPDATE_TIPS_JOB_SCHEDULE_INTERVAL = readEnvInt('UPDATE_TIPS_JOB_SCHEDULE_INTERVAL', 60 * 5)

/**
 * The interval (in seconds) to schedule the clean-old-jobs job.
 */
export const CLEAN_OLD_JOBS_INTERVAL = readEnvInt('CLEAN_OLD_JOBS_INTERVAL', 60 * 60 * 8)

/**
 * The default number of remote dumps to open when performing a global find-reference operation.
 */
export const DEFAULT_REFERENCES_NUM_REMOTE_DUMPS = readEnvInt('DEFAULT_REFERENCES_NUM_REMOTE_DUMPS', 10)

/**
 * The interval (in seconds) to schedule the clean-failed-jobs job.
 */
export const CLEAN_FAILED_JOBS_INTERVAL = readEnvInt('CLEAN_FAILED_JOBS_INTERVAL', 60 * 60 * 8)

/**
 * The default page size for the job endpoints.
 */
export const DEFAULT_JOB_PAGE_SIZE = readEnvInt('DEFAULT_JOB_PAGE_SIZE', 50)

/**
 * The maximum number of jobs to search in one call to the search-jobs.lua script.
 */
export const MAX_JOB_SEARCH = readEnvInt('MAX_JOB_SEARCH', 10000)

/**
 * The default page size for the dumps endpoint.
 */
export const DEFAULT_DUMP_PAGE_SIZE = readEnvInt('DEFAULT_DUMP_PAGE_SIZE', 50)
