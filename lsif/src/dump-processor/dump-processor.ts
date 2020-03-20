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
import { waitForConfiguration } from '../shared/config/config'
import { UploadManager } from '../shared/store/uploads'
import * as pgModels from '../shared/models/pg'
import { convertDatabase } from './conversion/conversion'
import { pick } from 'lodash'
import AsyncPolling from 'async-polling'
import { DumpManager } from '../shared/store/dumps'
import { DependencyManager } from '../shared/store/dependencies'
import { EntityManager } from 'typeorm'
import { SRC_FRONTEND_INTERNAL } from '../shared/config/settings'
import { updateCommitsAndDumpsVisibleFromTip } from '../shared/visibility'
import { startExpressApp } from '../shared/api/init'

/**
 * Runs the processor that converts LSIF uploads.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-dump-processor', fetchConfiguration())

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create database connection and entity wrapper classes
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const dumpManager = new DumpManager(connection)
    const uploadManager = new UploadManager(connection)
    const dependencyManager = new DependencyManager(connection)

    // Start metrics server
    startExpressApp({ routes: [], port: settings.METRICS_PORT, logger })

    const convert = async (upload: pgModels.LsifUpload, entityManager: EntityManager): Promise<void> => {
        logger.debug('Selected upload to convert', { uploadId: upload.id })

        let span: Span | undefined
        if (tracer) {
            // Extract tracing context from upload
            const publisher = tracer.extract(FORMAT_TEXT_MAP, JSON.parse(upload.tracingContext))
            span = tracer.startSpan(
                'Upload selected for conversion',
                publisher ? { references: [followsFrom(publisher)] } : {}
            )
        }

        // Tag tracing context with uploadId and arguments
        const ctx = addTags({ logger, span }, { uploadId: upload.id, ...pick(upload, 'repository', 'commit', 'root') })

        await instrument(
            metrics.uploadConversionDurationHistogram,
            metrics.uploadConversionDurationErrorsCounter,
            (): Promise<void> =>
                logAndTraceCall(ctx, 'Converting upload', async (ctx: TracingContext) => {
                    // Convert the database and populate the cross-dump package data
                    await convertDatabase(
                        entityManager,
                        dumpManager,
                        dependencyManager,
                        SRC_FRONTEND_INTERNAL,
                        upload,
                        ctx
                    )

                    // Remove overlapping dumps that would cause a unique index error once this upload has
                    // transitioned into the completed state. As this is done in a transaction, we do not
                    // delete the files on disk right away. These files will be cleaned up by a dump
                    // processor in a future cleanup task.
                    await dumpManager.deleteOverlappingDumps(
                        upload.repositoryId,
                        upload.commit,
                        upload.root,
                        upload.indexer,
                        { logger, span },
                        entityManager
                    )

                    // Update the conversion state after we've written the dump database file as the
                    // next step assumes that the processed upload is present in the dumps views. The
                    // remainder of the task may still fail, in which case the entire transaction is
                    // rolled back, so we don't want to commit yet.
                    await uploadManager.markComplete(upload, entityManager)

                    // Update visibility flag for this repository.
                    await updateCommitsAndDumpsVisibleFromTip({
                        entityManager,
                        dumpManager,
                        frontendUrl: SRC_FRONTEND_INTERNAL,
                        repositoryId: upload.repositoryId,
                        commit: upload.commit,
                        ctx,
                    })
                })
        )
    }

    logger.debug('Polling database for unconverted uploads')

    AsyncPolling(async end => {
        while (await uploadManager.dequeueAndConvert(convert, logger)) {
            // Immediately poll again if we converted an upload
        }

        end()
    }, settings.POLLING_INTERVAL * 1000).run()
}

// Initialize logger
const appLogger = createLogger('lsif-dump-processor')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
