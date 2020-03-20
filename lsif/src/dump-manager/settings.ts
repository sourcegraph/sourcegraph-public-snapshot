import { readEnvInt } from '../shared/settings'

/** Which port to run the LSIF server on. Defaults to 3187. */
export const HTTP_PORT = readEnvInt('HTTP_PORT', 3187)

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

/** The maximum rate that the server will send upload payloads. */
export const MAXIMUM_SERVE_BITS_PER_SECOND = readEnvInt('MAXIMUM_SERVE_BITS_PER_SECOND', 1024 * 1024 * 1024 * 10)

/** The maximum chunksize the server will use to send upload payloads. */
export const MAXIMUM_SERVE_CHUNKSIZE = readEnvInt('MAXIMUM_SERVE_CHUNKSIZE', 1024 * 1024 * 10)

/** The maximum rate that the server will receive upload payloads. */
export const MAXIMUM_UPLOAD_BITS_PER_SECOND = readEnvInt('MAXIMUM_UPLOAD_BITS_PER_SECOND', 1024 * 1024 * 1024 * 10)

/** The maximum chunksize the server will use to receive upload payloads. */
export const MAXIMUM_UPLOAD_CHUNKSIZE = readEnvInt('MAXIMUM_DOWNLOAD_CHUNKSIZE', 1024 * 1024 * 10)
