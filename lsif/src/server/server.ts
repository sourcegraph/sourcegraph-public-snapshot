import * as constants from '../constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from './backend/cache.settings'
import cors from 'cors'
import express from 'express'
import promClient from 'prom-client'
import { Backend } from './backend/backend'
import { Connection } from 'typeorm'
import { createDumpRouter } from './routes/dumps'
import { createJobRouter } from './routes/jobs'
import { createLogger } from '../shared/logging'
import { createLsifRouter } from './routes/lsif'
import { createMetaRouter } from './routes/meta'
import { createPostgresConnection } from '../shared/database/postgres'
import { createQueue, ensureOnlyRepeatableJob } from '../shared/queue/queue'
import { createTracer } from '../shared/tracing'
import { dbFilename, dbFilenameOld, ensureDirectory } from '../shared/util'
import { default as tracingMiddleware } from 'express-opentracing'
import { defineRedisCommands } from './redis/redis'
import { errorHandler } from './middleware/errors'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { LsifDump } from '../shared/models/xrepo'
import { metricsMiddleware } from './middleware/metrics'
import { pipeline as _pipeline } from 'stream'
import { queueSizeGauge } from './metrics'
import { waitForConfiguration } from '../shared/config/config'
import { XrepoDatabase } from '../shared/xrepo/xrepo'
import {
    STORAGE_ROOT,
    REDIS_ENDPOINT,
    UPDATE_TIPS_JOB_SCHEDULE_INTERVAL,
    CLEAN_OLD_JOBS_INTERVAL,
    CLEAN_FAILED_JOBS_INTERVAL,
    HTTP_PORT,
} from './settings'
import { connectionCacheCapacityGauge, documentCacheCapacityGauge, resultChunkCacheCapacityGauge } from './metrics'

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
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
    connectionCacheCapacityGauge.set(settings.CONNECTION_CACHE_CAPACITY)
    documentCacheCapacityGauge.set(settings.DOCUMENT_CACHE_CAPACITY)
    resultChunkCacheCapacityGauge.set(settings.RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(STORAGE_ROOT, connection)
    const backend = new Backend(STORAGE_ROOT, xrepoDatabase, fetchConfiguration)

    // Temporary migrations
    await moveDatabaseFilesToSubdir() // TODO - remove after 3.12
    await ensureFilenamesAreIDs(connection) // TODO - remove after 3.10

    // Create queue to publish convert
    const queue = createQueue(REDIS_ENDPOINT, logger)

    // Schedule jobs on timers
    await ensureOnlyRepeatableJob(queue, 'update-tips', {}, UPDATE_TIPS_JOB_SCHEDULE_INTERVAL * 1000)
    await ensureOnlyRepeatableJob(queue, 'clean-old-jobs', {}, CLEAN_OLD_JOBS_INTERVAL * 1000)
    await ensureOnlyRepeatableJob(queue, 'clean-failed-jobs', {}, CLEAN_FAILED_JOBS_INTERVAL * 1000)

    // Update queue size metric on a timer
    setInterval(
        () =>
            queue
                .count()
                .then(count => queueSizeGauge.set(count))
                .catch(() => {}),
        1000
    )

    // Register the required commands on the queue's Redis client
    const scriptedClient = await defineRedisCommands(queue.client)

    const app = express()
    app.use(cors())

    if (tracer !== undefined) {
        app.use(tracingMiddleware({ tracer }))
    }

    app.use(
        loggingMiddleware({
            winstonInstance: logger,
            level: 'debug',
            ignoredRoutes: ['/ping', '/healthz', '/metrics'],
            requestWhitelist: ['method', 'url', 'query'],
            msg: 'request',
        })
    )
    app.use(metricsMiddleware)

    // Register endpoints
    app.use(createMetaRouter())
    app.use(createLsifRouter(backend, queue, logger, tracer))
    app.use(createDumpRouter(backend, logger, tracer))
    app.use(createJobRouter(queue, scriptedClient, logger, tracer))

    // Error handler must be registered last
    app.use(errorHandler(logger))

    app.listen(HTTP_PORT, () => logger.debug('listening', { port: HTTP_PORT }))
}

/**
 * Move all db files in storage root to a subdirectory.
 */
async function moveDatabaseFilesToSubdir(): Promise<void> {
    for (const filename of await fs.readdir(STORAGE_ROOT)) {
        if (filename.endsWith('.db')) {
            await fs.rename(path.join(STORAGE_ROOT, filename), path.join(STORAGE_ROOT, constants.DBS_DIR, filename))
        }
    }
}

/**
 * If it hasn't been done already, migrate from the old pre-3.9 filename format
 * `$REPO@$COMMIT.lsif.db` to the new format `$ID.lsif.db`.
 */
async function ensureFilenamesAreIDs(db: Connection): Promise<void> {
    const doneFile = path.join(STORAGE_ROOT, 'id-based-filenames')
    if (await fs.exists(doneFile)) {
        // Already migrated.
        return
    }

    for (const dump of await db.getRepository(LsifDump).find()) {
        const oldFile = dbFilenameOld(STORAGE_ROOT, dump.repository, dump.commit)
        const newFile = dbFilename(STORAGE_ROOT, dump.id, dump.repository, dump.commit)
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
