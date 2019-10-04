import * as fs from 'mz/fs'
import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import promBundle from 'express-prom-bundle'
import uuid from 'uuid'
import { convertLsif } from './importer'
import { createDatabaseFilename, ensureDirectory, readEnvInt } from './util'
import { createLogger } from './logging'
import { createPostgresConnection } from './connection'
import { JobsHash, Worker } from 'node-resque'
import { Logger } from 'winston'
import { XrepoDatabase } from './xrepo'
import { Tracer } from 'opentracing'
import { MonitoringContext, monitor } from './monitoring'
import { waitForConfiguration, ConfigurationFetcher } from './config'
import { createTracer } from './tracing'
import { updateCommits } from './commits'

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
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param configurationFetcher A function that returns the current configuration.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
function createConvertJob(
    xrepoDatabase: XrepoDatabase,
    configurationFetcher: ConfigurationFetcher,
    logger: Logger,
    tracer: Tracer | undefined
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        const ctx = {
            logger: logger.child({ jobId: uuid.v4(), repository, commit }),
            span: tracer && tracer.startSpan('create convert job'),
        }

        await monitor(ctx, 'converting LSIF data', async (ctx: MonitoringContext) => {
            const input = fs.createReadStream(filename)
            const tempFile = path.join(STORAGE_ROOT, 'tmp', uuid.v4())

            try {
                // Create database in a temp path
                const { packages, references } = await convertLsif(input, tempFile, ctx)

                // Move the temp file where it can be found by the server
                await fs.rename(tempFile, createDatabaseFilename(STORAGE_ROOT, repository, commit))

                // Add the new database to the xrepo db
                await monitor(ctx, 'populating cross-repo database', () =>
                    xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
                )
            } catch (e) {
                // Don't leave busted artifacts
                await fs.unlink(tempFile)
                throw e
            }

            // Update commit parentage information for this commit
            await updateCommits(configurationFetcher().gitServers, xrepoDatabase, repository, commit, ctx)

            // Remove input
            await fs.unlink(filename)
        })
    }
}

/**
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Read configuration from frontend
    const configurationFetcher = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-worker', configurationFetcher())

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await ensureDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create cross-repo database
    const connection = await createPostgresConnection(configurationFetcher(), logger)
    const xrepoDatabase = new XrepoDatabase(connection)

    // Start metrics server
    startMetricsServer(logger)

    // Create worker and start processing jobs
    await startWorker(logger, {
        convert: createConvertJob(xrepoDatabase, configurationFetcher, logger, tracer),
    })
}

/**
 * Connect to redis and begin processing work with the given hash of job functions.
 *
 * @param logger The logger instance.
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
    worker.on('error', error => logger.error('worker error', { error }))

    await worker.connect()
    exitHook(() => worker.end())
    await worker.start()
}

/**
 * Create an express server that only has /healthz and /metric endpoints.
 *
 * @param logger The logger instance.
 */
function startMetricsServer(logger: Logger): void {
    const app = express()
    app.get('/healthz', (_, res) => res.send('ok'))
    app.use(promBundle({}))

    app.listen(WORKER_METRICS_PORT, () => logger.debug('listening', { port: WORKER_METRICS_PORT }))
}

// Initialize logger
const appLogger = createLogger('lsif-worker')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
