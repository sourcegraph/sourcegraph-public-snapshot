import * as settings from '../settings'
import { updateQueueSizeGauge, resetStalledUploads, cleanOldUploads } from './uploads'
import { Connection } from 'typeorm'
import { Logger } from 'winston'
import { UploadManager } from '../../shared/store/uploads'
import { DumpManager } from '../../shared/store/dumps'
import { TaskRunner } from '../../shared/tasks'

/**
 * Begin running cleanup tasks on a schedule in the background.
 *
 * @param connection The Postgres connection.
 * @param dumpManager The dumps manager instance.
 * @param uploadManager The uploads manager instance.
 * @param logger The logger instance.
 */
export function startTasks(
    connection: Connection,
    dumpManager: DumpManager,
    uploadManager: UploadManager,
    logger: Logger
): void {
    const runner = new TaskRunner(connection, logger)

    runner.register('Resetting stalled uploads', settings.RESET_STALLED_UPLOADS_INTERVAL, ctx =>
        resetStalledUploads(uploadManager, ctx)
    )

    runner.register('Cleaning old uploads', settings.CLEAN_OLD_UPLOADS_INTERVAL, ctx =>
        cleanOldUploads(uploadManager, ctx)
    )

    runner.register(
        'Updating metrics',
        settings.UPDATE_QUEUE_SIZE_GAUGE_INTERVAL,
        (): Promise<void> => updateQueueSizeGauge(uploadManager),
        true
    )

    runner.run()
}
