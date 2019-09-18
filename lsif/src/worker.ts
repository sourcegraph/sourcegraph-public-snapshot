import * as fs from 'mz/fs'
import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import promBundle from 'express-prom-bundle'
import uuid from 'uuid'
import { ConnectionCache } from './cache'
import { connectionCacheCapacityGauge } from './metrics'
import { convertLsif } from './importer'
import { createDatabaseFilename, ensureDirectory, readEnvInt } from './util'
import { createLogger } from './logger'
import { JobsHash, Worker } from 'node-resque'
import { Logger } from 'winston'
import { XrepoDatabase } from './xrepo'

/**
 * Which port to run the worker metrics server on. Defaults to 3187.
 */
const WORKER_METRICS_PORT = readEnvInt('WORKER_METRICS_PORT', 3187)

/**
 * The host and port running the redis instance containing work queues.
 *
 * Set addresses. Prefer in this order:
 *   - Specific envvar REDIS_STORE_ENDPOINT
 *   - Fallback envvar REDIS_ENDPOINT
 *   - redis-store:6379
 *
 *  Additionally keep this logic in sync with pkg/redispool/redispool.go and cmd/server/redis.go
 */
const REDIS_ENDPOINT = process.env.REDIS_STORE_ENDPOINT || process.env.REDIS_ENDPOINT || 'redis-store:6379'

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 100)

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 *
 * @param logger The application logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Update cache capacities on startup
    connectionCacheCapacityGauge.set(CONNECTION_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await ensureDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create backend
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const filename = path.join(STORAGE_ROOT, 'xrepo.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)

    const jobFunctions = {
        convert: createConvertJob(xrepoDatabase, logger),
    }

    // Start metrics server
    startMetricsServer(logger)

    // Create worker and start processing jobs
    await startWorker(logger, jobFunctions)
}

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param logger The worker's logger instance.
 */
function createConvertJob(
    xrepoDatabase: XrepoDatabase,
    logger: Logger
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        const jobLogger = logger.child({ jobId: uuid.v4(), repository, commit })
        const jobTimer = jobLogger.startTimer()
        jobLogger.info('converting LSIF data')

        const input = fs.createReadStream(filename)
        const tempFile = path.join(STORAGE_ROOT, 'tmp', uuid.v4())

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile, jobLogger)

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, createDatabaseFilename(STORAGE_ROOT, repository, commit))

            // Add the new database to the xrepo db
            const xrepoTimer = jobLogger.startTimer()
            jobLogger.debug('populating cross-repo database')
            await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
            xrepoTimer.done({ message: 'populated cross-repo database', level: 'debug' })
        } catch (e) {
            jobLogger.error('failed to convert LSIF data', { error: e })
            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }

        jobTimer.done({ message: 'converted LSIF data', level: 'info' })

        // Remove input
        await fs.unlink(filename)
    }
}

/**
 * Connect to redis and begin processing work with the given hash of job functions.
 *
 * @param logger The worker's logger instance.
 * @param jobFunctions An object whose values are the functions to execute for a job name matching its key.
 */
async function startWorker(
    logger: Logger,
    jobFunctions: { [name: string]: (...args: any[]) => Promise<any> }
): Promise<void> {
    const [host, port] = REDIS_ENDPOINT.split(':', 2)

    const connectionOptions = {
        host,
        port: parseInt(port, 10),
        namespace: 'lsif',
    }

    const jobs: JobsHash = {}
    for (const key of Object.keys(jobFunctions)) {
        jobs[key] = { perform: jobFunctions[key] }
    }

    // Create worker and log the interesting events
    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, jobs)
    worker.on('start', () => logger.debug('worker started'))
    worker.on('end', () => logger.debug('worker ended'))
    worker.on('poll', () => logger.debug('worker polling queue'))
    worker.on('ping', () => logger.debug('worker pinging queue'))
    worker.on('job', (_, job) => logger.debug('worker accepted job', { job }))
    worker.on('success', (_, job, result) => logger.debug('worker completed job', { job, result }))
    worker.on('failure', (_, job, failure) => logger.debug('worker failed job', { job, failure }))
    worker.on('cleaning_worker', (worker, pid) =>
        logger.debug('worker cleaning old sibling', { worker: `${worker}:${pid}` })
    )
    worker.on('error', e => logger.error('worker error', { error: e }))

    await worker.connect()
    exitHook(() => worker.end())
    await worker.start()
}

/**
 * Create an express server that only has /healthz and /metric endpoints.
 *
 * @param logger The worker's logger instance.
 */
function startMetricsServer(logger: Logger): void {
    const app = express()
    app.get('/healthz', (_, res) => res.send('ok'))
    app.use(promBundle({}))

    app.listen(WORKER_METRICS_PORT, () =>
        logger.debug('listening', { port: WORKER_METRICS_PORT, foo: [1, 2, 3], bar: { logger } })
    )
}

// Initialize logger
const appLogger = createLogger('lsif-workers')

// Launch!
main(appLogger).catch(e => appLogger.error('failed to start process', { error: e }))
