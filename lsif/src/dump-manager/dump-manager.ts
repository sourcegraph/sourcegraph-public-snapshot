import * as constants from '../shared/constants'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { createLogger } from '../shared/logging'
import { ensureDirectory } from '../shared/paths'
import { Logger } from 'winston'
import { startExpressApp } from '../shared/api/init'
import { createDatabaseRouter } from './routes/database'
import { createPostgresConnection } from '../shared/database/postgres'
import { waitForConfiguration } from '../shared/config/config'
import { startTasks } from './tasks'

/**
 * No-op dump-manager process.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration(logger)

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create database connection
    const connection = await createPostgresConnection(fetchConfiguration(), logger)

    // Start background tasks
    startTasks(connection, logger)

    const routers = [createDatabaseRouter(logger)]

    // Start server
    startExpressApp({ port: settings.HTTP_PORT, routers, logger })
}

// Initialize logger
const appLogger = createLogger('lsif-dump-manager')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
