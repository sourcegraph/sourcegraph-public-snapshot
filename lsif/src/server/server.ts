import * as constants from '../shared/constants'
import * as fs from 'mz/fs'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import * as xrepoModels from '../shared/models/xrepo'
import express from 'express'
import promClient from 'prom-client'
import { Backend } from './backend/backend'
import { Connection } from 'typeorm'
import { createDumpRouter } from './routes/dumps'
import { createLogger } from '../shared/logging'
import { createLsifRouter } from './routes/lsif'
import { createMetaRouter } from './routes/meta'
import { createPostgresConnection } from '../shared/database/postgres'
import { createTracer } from '../shared/tracing'
import { createUploadRouter } from './routes/uploads'
import { dbFilename, dbFilenameOld, ensureDirectory } from '../shared/paths'
import { default as tracingMiddleware } from 'express-opentracing'
import { errorHandler } from './middleware/errors'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { metricsMiddleware } from './middleware/metrics'
import { startTasks } from './tasks/runner'
import { UploadsManager } from '../shared/uploads/uploads'
import { waitForConfiguration } from '../shared/config/config'
import { XrepoDatabase } from '../shared/xrepo/xrepo'

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
    const tracer = createTracer('lsif-server', fetchConfiguration())

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
    const xrepoDatabase = new XrepoDatabase(connection, settings.STORAGE_ROOT)
    const backend = new Backend(settings.STORAGE_ROOT, xrepoDatabase, fetchConfiguration)
    const uploadsManager = new UploadsManager(connection)

    // Temporary migrations
    await moveDatabaseFilesToSubdir() // TODO - remove after 3.12
    await ensureFilenamesAreIDs(connection) // TODO - remove after 3.10

    // Start background tasks
    startTasks(connection, uploadsManager, logger)

    const app = express()

    if (tracer !== undefined) {
        app.use(tracingMiddleware({ tracer }))
    }

    app.use(
        loggingMiddleware({
            winstonInstance: logger,
            level: 'debug',
            ignoredRoutes: ['/ping', '/healthz', '/metrics'],
            requestWhitelist: ['method', 'url'],
            msg: 'Handled request',
        })
    )
    app.use(metricsMiddleware)

    // Register endpoints
    app.use(createMetaRouter())
    app.use(createDumpRouter(backend))
    app.use(createUploadRouter(uploadsManager))
    app.use(createLsifRouter(backend, uploadsManager, logger, tracer))

    // Error handler must be registered last
    app.use(errorHandler(logger))

    app.listen(settings.HTTP_PORT, () => logger.debug('LSIF API server listening on', { port: settings.HTTP_PORT }))
}

/**
 * Move all db files in storage root to a subdirectory.
 */
async function moveDatabaseFilesToSubdir(): Promise<void> {
    for (const filename of await fs.readdir(settings.STORAGE_ROOT)) {
        if (filename.endsWith('.db')) {
            await fs.rename(
                path.join(settings.STORAGE_ROOT, filename),
                path.join(settings.STORAGE_ROOT, constants.DBS_DIR, filename)
            )
        }
    }
}

/**
 * If it hasn't been done already, migrate from the old pre-3.9 filename format
 * `$REPO@$COMMIT.lsif.db` to the new format `$ID.lsif.db`.
 */
async function ensureFilenamesAreIDs(db: Connection): Promise<void> {
    const doneFile = path.join(settings.STORAGE_ROOT, 'id-based-filenames')
    if (await fs.exists(doneFile)) {
        // Already migrated.
        return
    }

    for (const dump of await db.getRepository(xrepoModels.LsifDump).find()) {
        const oldFile = dbFilenameOld(settings.STORAGE_ROOT, dump.repository, dump.commit)
        const newFile = dbFilename(settings.STORAGE_ROOT, dump.id, dump.repository, dump.commit)
        if (!(await fs.exists(oldFile))) {
            continue
        }
        await fs.rename(oldFile, newFile)
    }

    // Create an empty done file to record that all files have been renamed.
    await fs.close(await fs.open(doneFile, 'w'))
}

// Initialize logger
const appLogger = createLogger('lsif-server')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
