import * as fs from 'mz/fs'
import * as path from 'path'
import express from 'express'
import promClient from 'prom-client'
import { convertLsif } from './importer'
import { dbFilename, ensureDirectory, readEnvInt } from './util'
import { createLogger } from './logging'
import { createPostgresConnection } from './connection'
import { Logger } from 'winston'
import { XrepoDatabase } from './xrepo'
import { Tracer, FORMAT_TEXT_MAP, Span, followsFrom } from 'opentracing'
import { createTracer, TracingContext, logAndTraceCall, addTags } from './tracing'
import { waitForConfiguration, ConfigurationFetcher } from './config'
import { jobDurationHistogram, jobDurationErrorsCounter } from './worker.metrics'
import { Job, JobStatusClean, Queue } from 'bull'
import { createQueue } from './queue'
import { instrument } from './metrics'
import { purgeOldDumps } from './retention'
import * as constants from './constants'

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
 * The maximum age (in seconds) that a job (completed or queued) will remain in redis.
 */
const JOB_MAX_AGE = readEnvInt('JOB_MAX_AGE', 60 * 60 * 24 * 7)

/**
 * The maximum age (in seconds) that the files for a failed job can remain on disk.
 */
const FAILED_JOB_MAX_AGE = readEnvInt('FAILED_JOB_MAX_AGE', 24 * 60 * 60)

/**
 * The maximum space (in bytes) that the dbs directory can use.
 */
const DBS_DIR_MAXIMUM_SIZE_BYTES = readEnvInt('DBS_DIR_MAXIMUM_SIZE_BYTES', 1024 * 1024 * 1024 * 10)

/**
 * Wrap a job processor with instrumentation.
 *
 * @param name The job name.
 * @param jobProcessor The job processor.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
const wrapJobProcessor = <T>(
    name: string,
    jobProcessor: (args: T, ctx: TracingContext) => Promise<void>,
    logger: Logger,
    tracer: Tracer | undefined
): ((job: Job) => Promise<void>) => async (job: Job) => {
    logger.debug(`${name} job accepted`, { jobId: job.id })

    // Destructure arguments and injected tracing context
    const { args, tracing } = job.data as { args: T; tracing: object }

    let span: Span | undefined
    if (tracer) {
        // Extract tracing context from job payload
        const publisher = tracer.extract(FORMAT_TEXT_MAP, tracing)
        span = tracer.startSpan(name, publisher ? { references: [followsFrom(publisher)] } : {})
    }

    // Tag tracing context with jobId and arguments
    const ctx = addTags({ logger, span }, { jobId: job.id, ...args })

    await instrument(
        jobDurationHistogram,
        jobDurationErrorsCounter,
        (): Promise<void> => logAndTraceCall(ctx, `${name} job`, (ctx: TracingContext) => jobProcessor(args, ctx))
    )
}

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 */
const createConvertJobProcessor = (xrepoDatabase: XrepoDatabase, fetchConfiguration: ConfigurationFetcher) => async (
    { repository, commit, root, filename }: { repository: string; commit: string; root: string; filename: string },
    ctx: TracingContext
): Promise<void> => {
    await logAndTraceCall(ctx, 'converting LSIF data', async (ctx: TracingContext) => {
        const input = fs.createReadStream(filename)
        const tempFile = path.join(STORAGE_ROOT, constants.TEMP_DIR, path.basename(filename))

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile, ctx)

            // Add packages and references to the xrepo db
            const dump = await logAndTraceCall(ctx, 'populating cross-repo database', () =>
                xrepoDatabase.addPackagesAndReferences(repository, commit, root, packages, references)
            )

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, dbFilename(STORAGE_ROOT, dump.id, repository, commit))
        } catch (e) {
            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }
    })

    // Update commit parentage information for this commit
    await xrepoDatabase.discoverAndUpdateCommit({
        repository,
        commit,
        gitserverUrls: fetchConfiguration().gitServers,
        ctx,
    })

    // Remove input
    await fs.unlink(filename)

    // Clean up disk space if necessary
    await purgeOldDumps(STORAGE_ROOT, xrepoDatabase, DBS_DIR_MAXIMUM_SIZE_BYTES, ctx)
}

const cleanStatuses: JobStatusClean[] = ['completed', 'wait', 'active', 'delayed', 'failed']

/*
 * Create a job that updates the tip of the default branch for every repository that has LSIF data.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 */
const createUpdateTipsJobProcessor = (xrepoDatabase: XrepoDatabase, fetchConfiguration: ConfigurationFetcher) => (
    args: { [K: string]: any },
    ctx: TracingContext
): Promise<void> =>
    xrepoDatabase.discoverAndUpdateTips({
        gitserverUrls: fetchConfiguration().gitServers,
        ctx,
    })

/**
 * Create a job that removes all job data older than `JOB_MAX_AGE`.
 *
 * @param queue The queue.
 * @param logger The logger instance.
 */
const createCleanOldJobsProcessor = (queue: Queue, logger: Logger) => async (
    _: {},
    ctx: TracingContext
): Promise<void> => {
    const removedJobs = await logAndTraceCall(ctx, 'cleaning old jobs', (ctx: TracingContext) =>
        Promise.all(cleanStatuses.map(status => queue.clean(JOB_MAX_AGE * 1000, status)))
    )

    const { logger: jobLogger = logger } = ctx

    for (const [status, count] of removedJobs.map((jobs, i) => [cleanStatuses[i], jobs.length])) {
        if (count > 0) {
            jobLogger.debug('cleaned old jobs', { status, count })
        }
    }
}

/**
 * Create a job that removes upload and temp files that are older than `FAILED_JOB_MAX_AGE`.
 * This assumes that a conversion job's total duration (from enqueue to completion) is less
 * than this interval during healthy operation.
 */
const createCleanFailedJobsProcessor = () => async (_: {}, ctx: TracingContext): Promise<void> => {
    await logAndTraceCall(ctx, 'cleaning failed jobs', async (ctx: TracingContext) => {
        const purgeFile = async (filename: string): Promise<void> => {
            const stat = await fs.stat(filename)
            if (Date.now() - stat.mtimeMs >= FAILED_JOB_MAX_AGE) {
                await fs.unlink(filename)
            }
        }

        for (const directory of [constants.TEMP_DIR, constants.UPLOADS_DIR]) {
            for (const filename of await fs.readdir(path.join(STORAGE_ROOT, directory))) {
                await purgeFile(path.join(STORAGE_ROOT, directory, filename))
            }
        }
    })
}

/**
 * Runs the worker which accepts LSIF conversion jobs from the work queue.
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
    await ensureDirectory(path.join(STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(STORAGE_ROOT, connection)

    // Start metrics server
    startMetricsServer(logger)

    // Create queue to poll for jobs
    const queue = createQueue(REDIS_ENDPOINT, logger)

    const convertJobProcessor = wrapJobProcessor(
        'convert',
        createConvertJobProcessor(xrepoDatabase, fetchConfiguration),
        logger,
        tracer
    )

    const updateTipsJobProcessor = wrapJobProcessor(
        'update-tips',
        createUpdateTipsJobProcessor(xrepoDatabase, fetchConfiguration),
        logger,
        tracer
    )

    const cleanOldJobsProcessor = wrapJobProcessor(
        'clean-old-jobs',
        createCleanOldJobsProcessor(queue, logger),
        logger,
        tracer
    )

    const cleanFailedJobsProcessor = wrapJobProcessor(
        'clean-failed-jobs',
        createCleanFailedJobsProcessor(),
        logger,
        tracer
    )

    // Start processing work
    queue.process('convert', convertJobProcessor).catch(() => {})
    queue.process('update-tips', updateTipsJobProcessor).catch(() => {})
    queue.process('clean-old-jobs', cleanOldJobsProcessor).catch(() => {})
    queue.process('clean-failed-jobs', cleanFailedJobsProcessor).catch(() => {})
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
