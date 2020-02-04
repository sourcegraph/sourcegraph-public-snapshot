import * as constants from '../shared/constants'
import * as fs from 'mz/fs'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import express from 'express'
import promClient from 'prom-client'
import { createClient } from 'redis'
import { promisify } from 'util'
import { Backend } from './backend/backend'
import { createLogger } from '../shared/logging'
import { createLsifRouter } from './routes/lsif'
import { createMetaRouter } from './routes/meta'
import { createPostgresConnection } from '../shared/database/postgres'
import { createTracer } from '../shared/tracing'
import { createUploadRouter } from './routes/uploads'
import { dbFilename, ensureDirectory, idFromFilename } from '../shared/paths'
import { default as tracingMiddleware } from 'express-opentracing'
import { errorHandler } from './middleware/errors'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { metricsMiddleware } from './middleware/metrics'
import { startTasks } from './tasks/runner'
import { UploadManager } from '../shared/store/uploads'
import { waitForConfiguration } from '../shared/config/config'
import { DumpManager } from '../shared/store/dumps'
import { DependencyManager } from '../shared/store/dependencies'
import { SRC_FRONTEND_INTERNAL } from '../shared/config/settings'

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
    const dumpManager = new DumpManager(connection)
    const uploadManager = new UploadManager(connection)
    const dependencyManager = new DependencyManager(connection)
    const backend = new Backend(settings.STORAGE_ROOT, dumpManager, dependencyManager, SRC_FRONTEND_INTERNAL)

    // Temporary migration
    await migrateFilenames() // TODO - remove after 3.15
    await clearOldRedisData(logger) // TODO - remove after 3.15

    // Start background tasks
    startTasks(connection, dumpManager, uploadManager, logger)

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
    app.use(createUploadRouter(uploadManager))
    app.use(createLsifRouter(backend, uploadManager, logger, tracer))

    // Error handler must be registered last
    app.use(errorHandler(logger))

    app.listen(settings.HTTP_PORT, () => logger.debug('LSIF API server listening on', { port: settings.HTTP_PORT }))
}

/**
 * If it hasn't been done already, migrate from the old pre-3.13 filename format
 * `$ID-$REPO@$COMMIT.lsif.db` to the new format `$ID.lsif.db`.
 */
async function migrateFilenames(): Promise<void> {
    const doneFile = path.join(settings.STORAGE_ROOT, 'id-only-based-filenames')
    if (await fs.exists(doneFile)) {
        // Already migrated.
        return
    }

    for (const basename of await fs.readdir(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))) {
        const id = idFromFilename(basename)
        if (!id) {
            continue
        }

        await fs.rename(
            path.join(settings.STORAGE_ROOT, constants.DBS_DIR, basename),
            dbFilename(settings.STORAGE_ROOT, id)
        )
    }

    // Create an empty done file to record that all files have been renamed.
    await fs.close(await fs.open(doneFile, 'w'))
}

/**
 * Remove all old LSIF data from redis.
 */
async function clearOldRedisData(logger: Logger): Promise<void> {
    const script = `
        for i, key in ipairs(redis.call('keys', 'lsif:*')) do
            redis.call('del', key)
        end

        for i, key in ipairs(redis.call('keys', 'bull:*')) do
            redis.call('del', key)
        end
    `

    const url = process.env.REDIS_STORE_ENDPOINT || process.env.REDIS_ENDPOINT || 'redis-store:6379'
    const urlWithProtocol = url.includes('//') ? url : `redis://${url}`

    try {
        const client = createClient(urlWithProtocol)
        const evalAsync: (script: string, numArgs: number) => Promise<void> = promisify(client.eval).bind(client)
        await evalAsync(script, 0)
    } catch (err) {
        logger.warning('Failed to clean old LSIF data from redis-store', { error: err })
    }
}

// Initialize logger
const appLogger = createLogger('lsif-server')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
