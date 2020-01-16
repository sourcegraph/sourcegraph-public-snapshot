import * as settings from '../settings'
import AsyncPolling from 'async-polling'
import {
    updateQueueSizeGauge,
    resetStalledUploads,
    cleanOldUploads,
    cleanFailedUploads,
    purgeOldDumps,
} from './uploads'
import { Connection } from 'typeorm'
import { logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Logger } from 'winston'
import { tryWithLock } from '../../shared/store/locks'
import { UploadManager } from '../../shared/store/uploads'
import { DumpManager } from '../../shared/store/dumps'

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
    /**
     * Start invoking the given task on an interval.
     *
     * @param task The task to invoke.
     * @param intervalMs The interval between invocations.
     */
    const runTask = (task: () => Promise<void>, intervalMs: number): void => {
        AsyncPolling(async end => {
            await task()
            end()
        }, intervalMs * 1000).run()
    }

    /**
     * Each task is performed with an exclusive advisory lock in Postgres. If another
     * server is already running this task, then this server instance will skip the
     * attempt.
     *
     * @param name The task name. Used for logging the span and generating the lock id.
     * @param task The task function.
     */
    const wrapTask = (name: string, task: (ctx: TracingContext) => Promise<void>): (() => Promise<void>) => () =>
        tryWithLock(connection, name, () => logAndTraceCall({ logger }, name, task))

    runTask(
        wrapTask('Resetting stalled uploads', ctx => resetStalledUploads(uploadManager, ctx)),
        settings.RESET_STALLED_UPLOADS_INTERVAL
    )

    runTask(
        wrapTask('Cleaning old uploads', ctx => cleanOldUploads(uploadManager, ctx)),
        settings.CLEAN_OLD_UPLOADS_INTERVAL
    )

    runTask(
        wrapTask('Purging old dumps', ctx =>
            purgeOldDumps(connection, dumpManager, settings.STORAGE_ROOT, settings.DBS_DIR_MAXIMUM_SIZE_BYTES, ctx)
        ),
        settings.PURGE_OLD_DUMPS_INTERVAL
    )

    runTask(
        wrapTask('Cleaning failed uploads', ctx => cleanFailedUploads(ctx)),
        settings.CLEAN_FAILED_UPLOADS_INTERVAL
    )

    runTask((): Promise<void> => updateQueueSizeGauge(uploadManager), settings.UPDATE_QUEUE_SIZE_GAUGE_INTERVAL)
}
