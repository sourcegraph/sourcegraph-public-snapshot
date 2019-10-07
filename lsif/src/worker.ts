import * as fs from 'mz/fs'
import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import promBundle from 'express-prom-bundle'
import { omit } from 'lodash'
import uuid from 'uuid'
import { convertLsif } from './importer'
import { createDatabaseFilename, ensureDirectory, readEnvInt } from './util'
import { createLogger } from './logging'
import { createPostgresConnection } from './connection'
import { JobsHash, Worker } from 'node-resque'
import { Logger } from 'winston'
import { XrepoDatabase } from './xrepo'
import { Tracer, FORMAT_TEXT_MAP, Span } from 'opentracing'
import { MonitoringContext, monitor } from './monitoring'
import { createTracer } from './tracing'
import { waitForConfiguration, ConfigurationFetcher } from './config'
import { discoverAndUpdateCommit } from './commits'

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
 * A generic job. Takes a hash of arguments specific to the job type and
 * a monitoring context that is pulled from the job payload. This connects
 * the trace of this job with the trace of the publisher.
 */
type Job = (args: { [K: string]: any }, ctx: MonitoringContext) => Promise<void>

/**
 * Create a monitoring context from the logger and tracing span
 * tagged with the given values. Will attempt to pull the parent
 * span from the `tracing` value, if it was supplied with the
 * work request.
 *
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 * @param tracing The value of the injected parent span.
 * @param tags The tags to apply to the logger.
 */
const createMonitoringContext = (
    logger: Logger,
    tracer: Tracer | undefined,
    tracing: string,
    tags: { [K: string]: any }
): MonitoringContext => {
    let span: Span | undefined
    if (tracer && tracing !== '') {
        span = tracer.startSpan('job', { childOf: tracer.extract(FORMAT_TEXT_MAP, tracing)! })
    }

    return {
        logger: logger.child({ jobId: uuid.v4(), ...tags }),
        span,
    }
}

/**
 * Invoke the given job with a monitoring context pulled from the job
 * payload.
 *
 * @param job The job to wrap.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
const wrap = (
    job: Job,
    logger: Logger,
    tracer: Tracer | undefined
): ((args: { [K: string]: any }) => Promise<void>) => (args: { [K: string]: any }): Promise<void> => {
    const jobArgs = omit(args, 'tracing')
    return job(jobArgs, createMonitoringContext(logger, tracer, args.tracing, jobArgs))
}

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 */
const createConvertJob = (xrepoDatabase: XrepoDatabase, fetchConfiguration: ConfigurationFetcher) => async (
    args: { [K: string]: any },
    ctx: MonitoringContext
): Promise<void> => {
    // Destructure job arguments
    const { repository, commit, filename } = args

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
    })

    // Update commit parentage information for this commit
    await discoverAndUpdateCommit(xrepoDatabase, repository, commit, fetchConfiguration().gitServers, ctx)

    // Remove input
    await fs.unlink(filename)
}

/**
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-worker', fetchConfiguration())

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await ensureDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(connection)

    // Start metrics server
    startMetricsServer(logger)

    // Create worker and start processing jobs
    await startWorker(logger, {
        convert: wrap(createConvertJob(xrepoDatabase, fetchConfiguration), logger, tracer),
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

    const formatJob = (job: any): any => ({ class: job.class, args: job.args[0] })

    // Create worker and log the interesting events
    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, jobs)
    worker.on('start', () => logger.debug('worker started'))
    worker.on('end', () => logger.debug('worker ended'))
    worker.on('poll', () => logger.debug('worker polling queue'))
    worker.on('ping', () => logger.debug('worker pinging queue'))
    worker.on('job', (_, job) => logger.debug('worker accepted job', { job: formatJob(job) }))
    worker.on('success', (_, job, result) => logger.debug('worker completed job', { job: formatJob(job), result }))
    worker.on('failure', (_, job, failure) => logger.debug('worker failed job', { job: formatJob(job), failure }))
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
