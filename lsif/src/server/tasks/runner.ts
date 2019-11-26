import * as settings from '../settings'
import { cleanFailedUploads, cleanOldUploads, resetStalledUploads, updateUnconvertedUploadsGauge } from './uploads'
import { Connection } from 'typeorm'
import { logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Logger } from 'winston'
import { Queue } from '../../shared/uploads/uploads'
import { tryWithLock } from '../../shared/locks/locks'

/**
 * Begin running cleanup tasks on a schedule in the background.
 *
 * @param connection The Postgres connection.
 * @param queue The queue instance.
 * @param logger The logger instance.
 */
export function startTasks(connection: Connection, queue: Queue, logger: Logger): void {
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

    /**
     * Invoke the task function in a loop. This uses `setTimeout` recursively rather
     * than `setInterval` so we do not stack long running tasks when intervals are
     * configured to be small.
     *
     * @param task The task function.
     * @param intervalMs The interval (in milliseconds) between invocations.
     */
    const startTask = (task: () => Promise<void>, intervalMs: number): void => {
        const recur = async (): Promise<void> => {
            await task()
            setTimeout(() => {
                recur().catch(() => {})
            }, intervalMs)
        }

        recur().catch(() => {})
    }

    // Wrap tasks
    const updateUnconvertedUploadsGaugeTask = (): Promise<void> => updateUnconvertedUploadsGauge(queue)
    const resetStalledUploadsTask = wrapTask('Resetting stalled uploads', ctx => resetStalledUploads(queue, ctx))
    const cleanOldUploadsTask = wrapTask('Cleaning old uploads', ctx => cleanOldUploads(queue, ctx))
    const cleanFailedUploadsTask = wrapTask('Cleaning failed uploads', ctx => cleanFailedUploads(ctx))

    // Start tasks on intervals
    startTask(updateUnconvertedUploadsGaugeTask, settings.UPDATE_UNCONVERTED_UPLOADS_GAUGE_INTERVAL * 1000)
    startTask(resetStalledUploadsTask, settings.RESET_STALLED_UPLOADS_INTERVAL * 1000)
    startTask(cleanOldUploadsTask, settings.CLEAN_OLD_UPLOADS_INTERVAL * 1000)
    startTask(cleanFailedUploadsTask, settings.CLEAN_FAILED_UPLOADS_INTERVAL * 1000)
}
