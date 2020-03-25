import { readEnvInt } from '../shared/settings'

/** Which port to run the LSIF server on. Defaults to 3186. */
export const HTTP_PORT = readEnvInt('HTTP_PORT', 3186)

/** HTTP address for internal LSIF dump manager server. */
export const LSIF_DUMP_MANAGER_URL = process.env.LSIF_DUMP_MANAGER_URL || 'http://localhost:3187'

/** Where on the file system to temporarily store LSIF uploads. This need not be a persistent volume. */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/** The default number of results to return from the upload endpoints. */
export const DEFAULT_UPLOAD_PAGE_SIZE = readEnvInt('DEFAULT_UPLOAD_PAGE_SIZE', 50)

/** The default number of results to return from the dumps endpoint. */
export const DEFAULT_DUMP_PAGE_SIZE = readEnvInt('DEFAULT_DUMP_PAGE_SIZE', 50)

/** The default number of location results to return when performing a find-references operation. */
export const DEFAULT_REFERENCES_PAGE_SIZE = readEnvInt('DEFAULT_REFERENCES_PAGE_SIZE', 100)

/** The interval (in seconds) to invoke the updateQueueSizeGaugeInterval task. */
export const UPDATE_QUEUE_SIZE_GAUGE_INTERVAL = readEnvInt('UPDATE_QUEUE_SIZE_GAUGE_INTERVAL', 5)

/** The interval (in seconds) to run the resetStalledUploads task. */
export const RESET_STALLED_UPLOADS_INTERVAL = readEnvInt('RESET_STALLED_UPLOADS_INTERVAL', 60)

/** The maximum age (in seconds) that an upload can be unlocked and in the `processing` state. */
export const STALLED_UPLOAD_MAX_AGE = readEnvInt('STALLED_UPLOAD_MAX_AGE', 5)

/** The interval (in seconds) to invoke the cleanOldUploads task. */
export const CLEAN_OLD_UPLOADS_INTERVAL = readEnvInt('CLEAN_OLD_UPLOADS_INTERVAL', 60 * 60 * 8) // 8 hours

/** The maximum age (in seconds) that an upload (completed or queued) will remain in Postgres. */
export const UPLOAD_MAX_AGE = readEnvInt('UPLOAD_UPLOAD_AGE', 60 * 60 * 24 * 7) // 1 week
