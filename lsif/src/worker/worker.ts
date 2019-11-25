import * as constants from '../shared/constants'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { addTags, createTracer, logAndTraceCall, TracingContext } from '../shared/tracing'
import { createCleanFailedJobsProcessor } from './processors/clean-failed-jobs'
import { createCleanOldJobsProcessor } from './processors/clean-old-jobs'
import { createConvertJobProcessor } from './processors/convert'
import { createLogger } from '../shared/logging'
import { createPostgresConnection } from '../shared/database/postgres'
import { createQueue } from '../shared/queue/queue'
import { ensureDirectory } from '../shared/paths'
import { followsFrom, FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrumentWithLabels } from '../shared/metrics'
import { Job } from 'bull'
import { Logger } from 'winston'
import { startMetricsServer } from './server'
import { waitForConfiguration } from '../shared/config/config'
import { XrepoDatabase } from '../shared/xrepo/xrepo'

/**
 * Wrap a job processor with instrumentation.
 *
 * @param type The job name.
 * @param jobProcessor The job processor.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
const wrapJobProcessor = <T>(
    type: string,
    jobProcessor: (job: Job, args: T, ctx: TracingContext) => Promise<void>,
    logger: Logger,
    tracer: Tracer | undefined
): ((job: Job) => Promise<void>) => async (job: Job) => {
    logger.debug(`Accepted ${type} job`, { jobId: job.id })

    // Destructure arguments and injected tracing context
    const { args, tracing }: { args: T; tracing: object } = job.data

    let span: Span | undefined
    if (tracer) {
        // Extract tracing context from job payload
        const publisher = tracer.extract(FORMAT_TEXT_MAP, tracing)
        span = tracer.startSpan(type, publisher ? { references: [followsFrom(publisher)] } : {})
    }

    // Tag tracing context with jobId and arguments
    const ctx = addTags({ logger, span }, { jobId: job.id, ...args })

    await instrumentWithLabels(
        metrics.jobDurationHistogram,
        metrics.jobDurationErrorsCounter,
        { class: type },
        (): Promise<void> =>
            logAndTraceCall(ctx, `Running ${type} job`, (ctx: TracingContext) => jobProcessor(job, args, ctx))
    )
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
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(connection, settings.STORAGE_ROOT)

    // Start metrics server
    startMetricsServer(logger)

    // Create queue to poll for jobs
    const queue = createQueue(settings.REDIS_ENDPOINT, logger)

    const convertJobProcessor = wrapJobProcessor(
        'convert',
        createConvertJobProcessor(connection, xrepoDatabase, fetchConfiguration),
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
    queue.process('clean-old-jobs', cleanOldJobsProcessor).catch(() => {})
    queue.process('clean-failed-jobs', cleanFailedJobsProcessor).catch(() => {})
}

// Initialize logger
const appLogger = createLogger('lsif-worker')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
