import * as constants from '../shared/constants'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { addTags, createTracer, logAndTraceCall, TracingContext } from '../shared/tracing'
import { createLogger } from '../shared/logging'
import { createPostgresConnection } from '../shared/database/postgres'
import { ensureDirectory } from '../shared/paths'
import { followsFrom, FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrument } from '../shared/metrics'
import { Logger } from 'winston'
import { startMetricsServer } from './server'
import { waitForConfiguration } from '../shared/config/config'
import { XrepoDatabase } from '../shared/xrepo/xrepo'
import * as xrepoModels from '../shared/models/xrepo'
import { Queue } from '../shared/uploads/uploads'
import { createUploadConverter, uploadConverter } from './conversion'

/**
 * Wrap an upload converter with logging and tracing.
 *
 * @param convert The upload converter.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
const wrapUploadConverter = (convert: uploadConverter, logger: Logger, tracer?: Tracer) => async (
    upload: xrepoModels.LsifUpload
): Promise<void> => {
    let span: Span | undefined
    if (tracer) {
        // Extract tracing context from upload record
        const publisher = tracer.extract(FORMAT_TEXT_MAP, {}) // TODO - from upload
        span = tracer.startSpan('convert', publisher ? { references: [followsFrom(publisher)] } : {})
    }

    const args = {
        repository: upload.repository,
        commit: upload.commit,
        root: upload.root,
        filename: upload.filename,
    }

    // Tag tracing context with uploadId and arguments
    const ctx = addTags({ logger, span }, { uploadId: upload.id, ...args })

    await instrument(metrics.uploadConversionDurationHistogram, metrics.uploadConversionDurationErrorsCounter, () =>
        logAndTraceCall(ctx, 'Converting upload', (ctx: TracingContext) => convert(args, upload.uploadedAt, ctx))
    )
}

/**
 * Runs the worker which converts raw LSIF uploads into SQLite databases.
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

    // Create cross-repo database, queue, and upload converter
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(connection, settings.STORAGE_ROOT)
    const queue = new Queue(connection)
    const convertUpload = wrapUploadConverter(
        createUploadConverter(connection, xrepoDatabase, fetchConfiguration),
        logger,
        tracer
    )

    // Start metrics server
    startMetricsServer(logger)

    // TODO - find a library that does this
    while (true) {
        logger.debug('Polling')

        // TODO - catch errors here
        const handled = await queue.dequeueAndConvert(convertUpload)
        if (!handled) {
            // TODO - configure polling interval
            // TODO - some kind of polling interval here
            await new Promise(r => setTimeout(r, 1000))
        }
    }
}

// Initialize logger
const appLogger = createLogger('lsif-worker')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
