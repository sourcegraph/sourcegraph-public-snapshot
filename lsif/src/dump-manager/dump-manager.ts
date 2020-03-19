import * as constants from '../shared/constants'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { createLogger } from '../shared/logging'
import { ensureDirectory } from '../shared/paths'
import { Logger } from 'winston'
import { startExpressApp } from '../shared/api/init'

/**
 * No-op dump-manager process.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // Start server
    startExpressApp({ routes: [], port: settings.HTTP_PORT, logger })
}

// Initialize logger
const appLogger = createLogger('lsif-dump-manager')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
