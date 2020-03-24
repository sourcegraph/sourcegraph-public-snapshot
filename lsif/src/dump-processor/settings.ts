import { readEnvInt } from '../shared/settings'

/** Which port to run the metrics server on. Defaults to 3188. */
export const METRICS_PORT = readEnvInt('METRICS_PORT', 3188)

/** HTTP address for internal LSIF dump manager server. */
export const LSIF_DUMP_MANAGER_URL = process.env.LSIF_DUMP_MANAGER_URL || 'http://localhost:3187'

/** Where on the file system to temporarily store LSIF uploads and SQLite files. This is NOT a persistent volume. */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/** The interval (in seconds) to poll the database for unconverted uploads. */
export const POLLING_INTERVAL = readEnvInt('POLLING_INTERVAL', 1)

/**
 * The target results per result chunk. This is used to determine the number of chunks
 * created during conversion, but does not guarantee that the distribution of hash keys
 * will wbe even. In practice, chunks are fairly evenly filled.
 */
export const RESULTS_PER_RESULT_CHUNK = readEnvInt('RESULTS_PER_RESULT_CHUNK', 500)

/** The maximum number of result chunks that will be created during conversion. */
export const MAX_NUM_RESULT_CHUNKS = readEnvInt('MAX_NUM_RESULT_CHUNKS', 1000)
