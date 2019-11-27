import * as constants from '../shared/constants'
import * as metrics from './metrics'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { addTags, createTracer, logAndTraceCall, TracingContext } from '../shared/tracing'
import { createLogger } from '../shared/logging'
import { createPostgresConnection } from '../shared/database/postgres'
import { ensureDirectory } from '../shared/paths'
import { Span } from 'opentracing'
import { instrument } from '../shared/metrics'
import { Logger } from 'winston'
import { startMetricsServer } from './server'
import { waitForConfiguration } from '../shared/config/config'
import { XrepoDatabase } from '../shared/xrepo/xrepo'
import { UploadsManager } from '../shared/uploads/uploads'
import * as xrepoModels from '../shared/models/xrepo'
import { convertUpload } from './processors/convert'
import { pick } from 'lodash'

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

    // Create database connection and entity wrapper classes
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(connection, settings.STORAGE_ROOT)
    const uploadsManager = new UploadsManager(connection)

    // Start metrics server
    startMetricsServer(logger)

    // TODO - find a library that does this
    while (true) {
        logger.debug('Polling')

        const handled = await uploadsManager.dequeueAndConvert(
            async (upload: xrepoModels.LsifUpload): Promise<void> => {
                logger.debug('Selected upload to convert', { uploadId: upload.id })

                let span: Span | undefined
                if (tracer) {
                    // TODO - pull this from upload record
                    // Extract tracing context from job payload
                    //     const publisher = tracer.extract(FORMAT_TEXT_MAP, tracing)
                    //     span = tracer.startSpan(type, publisher ? { references: [followsFrom(publisher)] } : {})
                }

                // Tag tracing context with uploadId and arguments
                const ctx = addTags(
                    { logger, span },
                    { uploadId: upload.id, ...pick(upload, 'repository', 'commit', 'root') }
                )

                await instrument(
                    metrics.jobDurationHistogram,
                    metrics.jobDurationErrorsCounter,
                    (): Promise<void> =>
                        logAndTraceCall(ctx, 'Converting upload', (ctx: TracingContext) =>
                            convertUpload(
                                connection,
                                xrepoDatabase,
                                fetchConfiguration,
                                pick(upload, 'repository', 'commit', 'root', 'filename', 'uploadedAt'),
                                ctx
                            )
                        )
                )
            }
        )

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
