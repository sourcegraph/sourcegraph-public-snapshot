import { readEnvInt } from '../shared/settings'

/** Which port to run the metrics server on. Defaults to 3188. */
export const METRICS_PORT = readEnvInt('METRICS_PORT', 3188)

/** The interval (in seconds) to poll the database for unconverted uploads. */
export const POLLING_INTERVAL = readEnvInt('POLLING_INTERVAL', 1)

/** Where on the file system to store LSIF files. */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'
