import * as constants from '../shared/constants'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { addTags, createTracer, logAndTraceCall, TracingContext } from '../shared/tracing'
import { createLogger } from '../shared/logging'
import { createPostgresConnection } from '../shared/database/postgres'
import { ensureDirectory } from '../shared/paths'
import { Span, FORMAT_TEXT_MAP, followsFrom } from 'opentracing'
import { instrument } from '../shared/metrics'
import { Logger } from 'winston'
import { startMetricsServer } from './server'
import { waitForConfiguration } from '../shared/config/config'
import { UploadManager } from '../shared/store/uploads'
import * as pgModels from '../shared/models/pg'
import { convertUpload } from './conversion/conversion'
import { pick } from 'lodash'
import AsyncPolling from 'async-polling'
import { DumpManager } from '../shared/store/dumps'
import { DependencyManager } from '../shared/store/dependencies'

/**
 * Runs the worker that converts LSIF uploads.
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

    // Create database connection and entity wrapper classes
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const dumpManager = new DumpManager(connection, settings.STORAGE_ROOT)
    const uploadManager = new UploadManager(connection)
    const dependencyManager = new DependencyManager(connection)

    // Start metrics server
    startMetricsServer(logger)

    const convert = async (upload: pgModels.LsifUpload): Promise<void> => {
        logger.debug('Selected upload to convert', { uploadId: upload.id })

        let span: Span | undefined
        if (tracer) {
            // Extract tracing context from upload
            const publisher = tracer.extract(FORMAT_TEXT_MAP, JSON.parse(upload.tracingContext))
            span = tracer.startSpan(
                'Upload selected by worker',
                publisher ? { references: [followsFrom(publisher)] } : {}
            )
        }

        // Tag tracing context with uploadId and arguments
        const ctx = addTags({ logger, span }, { uploadId: upload.id, ...pick(upload, 'repository', 'commit', 'root') })

        await instrument(
            metrics.uploadConversionDurationHistogram,
            metrics.uploadConversionDurationErrorsCounter,
            (): Promise<void> =>
                logAndTraceCall(ctx, 'Converting upload', (ctx: TracingContext) =>
                    convertUpload(connection, dumpManager, dependencyManager, fetchConfiguration, upload, ctx)
                )
        )
    }

    logger.debug('Worker polling database for unconverted uploads')

    AsyncPolling(async end => {
        while (await uploadManager.dequeueAndConvert(convert, logger)) {
            // Immediately poll again if we converted an upload
        }

        end()
    }, settings.WORKER_POLLING_INTERVAL * 1000).run()
}

// Initialize logger
const appLogger = createLogger('lsif-worker')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
