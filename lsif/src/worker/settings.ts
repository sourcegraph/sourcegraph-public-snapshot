import { readEnvInt } from '../shared/settings'

/**
 * Which port to run the worker metrics server on. Defaults to 3187.
 */
export const WORKER_METRICS_PORT = readEnvInt('WORKER_METRICS_PORT', 3187)

/**
 * The interval (in seconds) to poll the database for unconverted uploads.
 */
export const WORKER_POLLING_INTERVAL = readEnvInt('WORKER_POLLING_INTERVAL', 1)

/**
 * Where on the file system to store LSIF files.
 */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * The maximum space (in bytes) that the dbs directory can use.
 */
export const DBS_DIR_MAXIMUM_SIZE_BYTES = readEnvInt('DBS_DIR_MAXIMUM_SIZE_BYTES', 1024 * 1024 * 1024 * 10)
