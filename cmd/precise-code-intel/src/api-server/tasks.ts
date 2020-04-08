import * as settings from './settings'
import { Connection } from 'typeorm'
import { Logger } from 'winston'
import { UploadManager } from '../shared/store/uploads'
import { ExclusivePeriodicTaskRunner } from '../shared/tasks'
import * as metrics from './metrics'
import { createSilentLogger } from '../shared/logging'
import { TracingContext } from '../shared/tracing'

/**
 * Begin running cleanup tasks on a schedule in the background.
 *
 * @param connection The Postgres connection.
 * @param uploadManager The uploads manager instance.
 * @param logger The logger instance.
 */
export function startTasks(connection: Connection, uploadManager: UploadManager, logger: Logger): void {
    const runner = new ExclusivePeriodicTaskRunner(connection, logger)

    runner.register({
        name: 'Updating metrics',
        intervalMs: settings.UPDATE_QUEUE_SIZE_GAUGE_INTERVAL,
        task: () => updateQueueSizeGauge(uploadManager),
        silent: true,
    })

    runner.register({
        name: 'Resetting stalled uploads',
        intervalMs: settings.RESET_STALLED_UPLOADS_INTERVAL,
        task: ({ ctx }) => resetStalledUploads(uploadManager, ctx),
    })

    runner.register({
        name: 'Cleaning old uploads',
        intervalMs: settings.CLEAN_OLD_UPLOADS_INTERVAL,
        task: ({ ctx }) => cleanOldUploads(uploadManager, ctx),
    })

    runner.run()
}

/**
 * Update the value of the unconverted uploads gauge.
 *
 * @param uploadManager The uploads manager instance.
 */
async function updateQueueSizeGauge(uploadManager: UploadManager): Promise<void> {
    metrics.unconvertedUploadSizeGauge.set(await uploadManager.getCount('queued'))
}

/**
 * Move all unlocked uploads that have been in `processing` state for longer than
 * `STALLED_UPLOAD_MAX_AGE` back to the `queued` state.
 *
 * @param uploadManager The uploads manager instance.
 * @param ctx The tracing context.
 */
async function resetStalledUploads(
    uploadManager: UploadManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> {
    for (const id of await uploadManager.resetStalled(settings.STALLED_UPLOAD_MAX_AGE)) {
        logger.debug('Reset stalled upload conversion', { id })
    }
}

/**
 * Remove all upload data older than `UPLOAD_MAX_AGE`.
 *
 * @param uploadManager The uploads manager instance.
 * @param ctx The tracing context.
 */
async function cleanOldUploads(
    uploadManager: UploadManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> {
    const count = await uploadManager.clean(settings.UPLOAD_MAX_AGE)
    if (count > 0) {
        logger.debug('Cleaned old uploads', { count })
    }
}
