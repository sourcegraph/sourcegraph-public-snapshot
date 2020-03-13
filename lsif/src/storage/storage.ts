import * as constants from '../shared/constants'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { createLogger } from '../shared/logging'
import { createTracer } from '../shared/tracing'
import { Logger } from 'winston'
import { waitForConfiguration } from '../shared/config/config'
import { makeExpressApp } from '../shared/api/init'
import express from 'express'
import { DumpManager } from '../shared/store/dumps'
import { createPostgresConnection } from '../shared/database/postgres'
import { ensureDirectory } from '../shared/paths'
import { createDatabaseRouter } from './routes/database'
import { jsonReplacer } from '../shared/encoding/json'

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
    const dumpManager = new DumpManager(connection)

    // // Update cache capacities on startup
    // metrics.connectionCacheCapacityGauge.set(settings.CONNECTION_CACHE_CAPACITY)
    // metrics.documentCacheCapacityGauge.set(settings.DOCUMENT_CACHE_CAPACITY)
    // metrics.resultChunkCacheCapacityGauge.set(settings.RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    // await ensureDirectory(settings.STORAGE_ROOT)
    // await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    // await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    // await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // // Create database connection and entity wrapper classes
    // const connection = await createPostgresConnection(fetchConfiguration(), logger)
    // const dumpManager = new DumpManager(connection)
    // const uploadManager = new UploadManager(connection)
    // const dependencyManager = new DependencyManager(connection)
    // const backend = new Backend(settings.STORAGE_ROOT, dumpManager, dependencyManager, SRC_FRONTEND_INTERNAL)

    // Run any app-level migrations. These migrations usually exist only
    // for a two-minor-version period in which we clean up old data and
    // fix outdated assumptions.
    //
    // These block the process from starting up until completion. Also
    // note that if the cleanup is handling an assumption from the last
    // minor version, there may be instances of that version running
    // after this migration step completes.
    // await migrate(connection, { logger })

    // Start background tasks
    // startTasks(connection, dumpManager, uploadManager, logger)

    const routes: express.Router[] = [createDatabaseRouter(dumpManager, logger)]

    // Register middleware and serve
    const app = makeExpressApp({ routes, logger, tracer /* , histogramSelector*/ })
    app.set('json replacer', jsonReplacer)
    app.listen(settings.HTTP_PORT, () => logger.debug('LSIF storage server listening on', { port: settings.HTTP_PORT }))
}

// function histogramSelector(route: string): promClient.Histogram<string> | undefined {
//     switch (route) {
//         case '/upload':
//             return metrics.httpUploadDurationHistogram

//         case '/exists':
//         case '/request':
//             return metrics.httpQueryDurationHistogram
//     }

//     return undefined
// }

// Initialize logger
const appLogger = createLogger('lsif-storage')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
