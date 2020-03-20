import { readEnvInt } from '../shared/settings'

/** Which port to run the LSIF server on. Defaults to 3187. */
export const HTTP_PORT = readEnvInt('HTTP_PORT', 3187)

/** HTTP address for internal LSIF HTTP API. */
export const LSIF_SERVER_URL = process.env.LSIF_SERVER_URL || 'http://lsif-server'

/** Where on the file system to store LSIF files. */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
export const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 100)

/** The maximum number of documents that can be held in memory at once. */
export const DOCUMENT_CACHE_CAPACITY = readEnvInt('DOCUMENT_CACHE_CAPACITY', 1024 * 1024 * 1024)

/** The maximum number of result chunks that can be held in memory at once. */
export const RESULT_CHUNK_CACHE_CAPACITY = readEnvInt('RESULT_CHUNK_CACHE_CAPACITY', 1024 * 1024 * 1024)

/** The interval (in seconds) to clean the dbs directory. */
export const PURGE_OLD_DUMPS_INTERVAL = readEnvInt('PURGE_OLD_DUMPS_INTERVAL', 60 * 30)

/** How many uploads to query at once when determining if a db file is unreferenced. */
export const DEAD_DUMP_CHUNK_SIZE = readEnvInt('DEAD_DUMP_CHUNK_SIZE', 100)

/** The maximum space (in bytes) that the dbs directory can use. */
export const DBS_DIR_MAXIMUM_SIZE_BYTES = readEnvInt('DBS_DIR_MAXIMUM_SIZE_BYTES', 1024 * 1024 * 1024 * 10)

/** The interval (in seconds) to invoke the cleanFailedUploads task. */
export const CLEAN_FAILED_UPLOADS_INTERVAL = readEnvInt('CLEAN_FAILED_UPLOADS_INTERVAL', 60 * 60 * 8)

/** The maximum age (in seconds) that the files for an unprocessed upload can remain on disk. */
export const FAILED_UPLOAD_MAX_AGE = readEnvInt('FAILED_UPLOAD_MAX_AGE', 24 * 60 * 60)

/** How many times to retry requests to lsif-server in the background. */
export const MAX_REQUEST_RETRIES = readEnvInt('MAX_REQUEST_RETRIES', 60)

/** How long to wait (in seconds) between lsif-server request attempts. */
export const REQUEST_RETRY_INTERVAL = readEnvInt('REQUEST_RETRY_INTERVAL', 5)
