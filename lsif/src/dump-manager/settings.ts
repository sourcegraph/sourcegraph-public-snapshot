import { readEnvInt } from '../shared/settings'

/** Which port to run the LSIF server on. Defaults to 3187. */
export const HTTP_PORT = readEnvInt('HTTP_PORT', 3187)

/** Where on the file system to store LSIF files. */
export const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'
