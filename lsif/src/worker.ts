import * as fs from 'mz/fs'
import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import promClient from 'prom-client'
import uuid from 'uuid'
import { convertLsif } from './importer'
import { dbFilename, ensureDirectory, readEnvInt } from './util'
import { createLogger } from './logging'
import { createPostgresConnection } from './connection'
import { JobsHash, Worker } from 'node-resque'
import { Logger } from 'winston'
import { XrepoDatabase } from './xrepo'
import { Tracer, FORMAT_TEXT_MAP, Span, followsFrom } from 'opentracing'
import { createTracer, TracingContext, logAndTraceCall, addTags } from './tracing'
import { waitForConfiguration, ConfigurationFetcher } from './config'
import { discoverAndUpdateCommit } from './commits'
import { jobDurationHistogram, jobEventsCounter } from './worker.metrics'

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
 * a tracing context that is pulled from the job payload. This connects
 * the trace of this job with the trace of the publisher.
 */
type Job = (args: object, ctx: TracingContext) => Promise<void>

/**
 * Create a tracing context from the logger and tracing span
 * tagged with the given values. Will attempt to pull the parent
 * span from the `tracing` value, if it was supplied with the
 * work request.
 *
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 * @param name The job name.
 * @param tracing The value of the injected parent span.
 * @param tags The tags to apply to the logger.
 */
const createTracingContext = (
    logger: Logger,
    tracer: Tracer | undefined,
    name: string,
    tracing: object,
    tags: { [K: string]: any }
): TracingContext => {
    let span: Span | undefined
    if (tracer) {
        const publisher = tracer.extract(FORMAT_TEXT_MAP, tracing)
        span = tracer.startSpan(name, publisher ? { references: [followsFrom(publisher)] } : {})
    }

    return addTags({ logger, span }, { jobId: uuid.v4(), ...tags })
}

/**
 * Invoke the given job with a tracing context pulled from the job
 * payload.
 *
 * @param name The job name.
 * @param job The job to wrap.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
const wrap = (
    name: string,
    job: Job,
    logger: Logger,
    tracer: Tracer | undefined
): (({ tracing, ...args }: object & { tracing: object }) => Promise<void>) => async ({
    tracing,
    ...args
}: object & { tracing: object }): Promise<void> => {
    const { logger: jobLogger, span = new Span() } = createTracingContext(logger, tracer, name, tracing, args)

    try {
        return await job(args, { logger: jobLogger, span })
    } finally {
        span.finish()
    }
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
    ctx: TracingContext
): Promise<void> => {
    // Destructure job arguments
    const { repository, commit, root, filename } = args as {
        repository: string
        commit: string
        root: string
        filename: string
    }

    await logAndTraceCall(ctx, 'converting LSIF data', async (ctx: TracingContext) => {
        const input = fs.createReadStream(filename)
        const tempFile = path.join(STORAGE_ROOT, 'tmp', uuid.v4())

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile, ctx)

            // Add packages and references to the xrepo db
            const dumpID = await logAndTraceCall(ctx, 'populating cross-repo database', () =>
                xrepoDatabase.addPackagesAndReferences(repository, commit, root, packages, references)
            )

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, dbFilename(STORAGE_ROOT, dumpID, repository, commit))
        } catch (e) {
            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }
    })

    // Update commit parentage information for this commit
    await discoverAndUpdateCommit({
        xrepoDatabase,
        repository,
        commit,
        gitserverUrls: fetchConfiguration().gitServers,
        ctx,
    })

    // Remove input
    await fs.unlink(filename)
}

/**
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

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
        convert: wrap('convert job', createConvertJob(xrepoDatabase, fetchConfiguration), logger, tracer),
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
    worker.on('cleaning_worker', (worker, pid) =>
        logger.debug('worker cleaning old sibling', { worker: `${worker}:${pid}` })
    )
    worker.on('error', error => logger.error('worker error', { error }))

    // Start a timer when accepting a job and end it when either
    // succeeding or failing. This is fine as we're not using a
    // multiWorker and only one job will be processed at a time.
    let end: (() => void) | undefined

    worker.on('job', (_, job) => {
        logger.debug('worker accepted job', { job: formatJob(job) })
        end = jobDurationHistogram.labels(job.class).startTimer()
    })

    worker.on('success', (_, job, result) => {
        logger.debug('worker completed job', { job: formatJob(job), result })
        jobEventsCounter.labels(job.class, 'success').inc()
        if (end) {
            end()
        }
    })

    worker.on('failure', (_, job, failure) => {
        logger.debug('worker failed job', { job: formatJob(job), failure })
        jobEventsCounter.labels(job.class, 'failure').inc()
        if (end) {
            end()
        }
    })

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
    app.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

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
