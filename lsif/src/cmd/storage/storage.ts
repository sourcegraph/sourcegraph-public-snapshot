import * as constants from '../../shared/constants'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { createLogger } from '../../shared/logging'
import { createTracer } from '../../shared/tracing'
import { Logger } from 'winston'
import { waitForConfiguration } from '../../shared/config/config'
import { makeExpressApp } from '../../shared/api/init'
import { createPostgresConnection } from '../../shared/database/postgres'
import { ensureDirectory } from '../../shared/paths'
import { createDatabaseRouter } from './routes/database'
import { createUploadRouter } from './routes/uploads'
import { startTasks } from './tasks'

/**
 * Runs the HTTP server that accepts LSIF dump uploads and responds to LSIF requests.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-storage', fetchConfiguration())

    // Update cache capacities on startup
    metrics.connectionCacheCapacityGauge.set(settings.CONNECTION_CACHE_CAPACITY)
    metrics.documentCacheCapacityGauge.set(settings.DOCUMENT_CACHE_CAPACITY)
    metrics.resultChunkCacheCapacityGauge.set(settings.RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create database connection and entity wrapper classes
    const connection = await createPostgresConnection(fetchConfiguration(), logger)

    // Start background tasks
    startTasks(connection, logger)

    // Register middleware and serve
    const app = makeExpressApp({
        routes: [createUploadRouter(logger), createDatabaseRouter(logger)],
        logger,
        tracer,
        histogramSelector,
    })

    app.listen(settings.HTTP_PORT, () => logger.debug('LSIF storage server listening on', { port: settings.HTTP_PORT }))
}

function histogramSelector(route: string): promClient.Histogram<string> | undefined {
    switch (route) {
        default:
        /* TODO */
    }

    return undefined
}

// Initialize logger
const appLogger = createLogger('lsif-storage')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
