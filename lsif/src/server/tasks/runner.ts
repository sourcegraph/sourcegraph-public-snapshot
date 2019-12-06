import * as settings from '../settings'
import AsyncPolling from 'async-polling'
import { cleanFailedJobs, cleanOldJobs, updateQueueSizeGauge } from './jobs'
import { Connection } from 'typeorm'
import { logAndTraceCall, TracingContext } from '../../shared/tracing'
import { Logger } from 'winston'
import { Queue } from 'bull'
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

    // Wrap tasks
    const updateQueueSizeGaugeTask = (): Promise<void> => updateQueueSizeGauge(queue)
    const cleanOldUploadsTask = wrapTask('Cleaning old uploads', ctx => cleanOldJobs(queue, ctx))
    const cleanFailedUploadsTask = wrapTask('Cleaning failed uploads', ctx => cleanFailedJobs(ctx))

    // Start tasks on intervals
    runTask(updateQueueSizeGaugeTask, settings.UPDATE_QUEUE_SIZE_GAUGE_INTERVAL)
    runTask(cleanOldUploadsTask, settings.CLEAN_OLD_JOBS_INTERVAL)
    runTask(cleanFailedUploadsTask, settings.CLEAN_FAILED_JOBS_INTERVAL)
}
