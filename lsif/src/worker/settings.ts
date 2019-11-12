import { readEnvInt } from '../shared/settings'

/**
 * Which port to run the worker metrics server on. Defaults to 3187.
 */
export const WORKER_METRICS_PORT = readEnvInt('WORKER_METRICS_PORT', 3187)

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
 * The maximum age (in seconds) that a job (completed or queued) will remain in redis.
 */
export const JOB_MAX_AGE = readEnvInt('JOB_MAX_AGE', 60 * 60 * 24 * 7)

/**
 * The maximum age (in seconds) that the files for a failed job can remain on disk.
 */
export const FAILED_JOB_MAX_AGE = readEnvInt('FAILED_JOB_MAX_AGE', 24 * 60 * 60)

/**
 * The maximum space (in bytes) that the dbs directory can use.
 */
export const DBS_DIR_MAXIMUM_SIZE_BYTES = readEnvInt('DBS_DIR_MAXIMUM_SIZE_BYTES', 1024 * 1024 * 1024 * 10)
