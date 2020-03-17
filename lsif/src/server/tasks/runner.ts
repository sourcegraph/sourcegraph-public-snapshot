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

interface Task {
    intervalMs: number
    handler: () => Promise<void>
}

/** A collection of tasks that are invoked periodically. */
export class TaskRunner {
    private tasks: Task[] = []

    /**
     * Create a new task runner.
     *
     * @param connection The Postgres connection.
     * @param logger The logger instance.
     */
    constructor(private connection: Connection, private logger: Logger) {}

    /**
     * Register a task to be performed while holding an exclusive advisory lock in Postgres.
     *
     * @param name The task name.
     * @param intervalMs The interval between task invocations.
     * @param task The function to invoke.
     */
    public register(name: string, intervalMs: number, task: (ctx: TracingContext) => Promise<void>): void {
        this.tasks.push({
            intervalMs,
            handler: () =>
                tryWithLock(this.connection, name, () => logAndTraceCall({ logger: this.logger }, name, task)),
        })
    }

    /** Start running all registered tasks on the specified interval. */
    public run(): void {
        for (const { intervalMs, handler } of this.tasks) {
            const fn = async (end: () => void): Promise<void> => {
                await handler()
                end()
            }

            AsyncPolling(fn, intervalMs * 1000).run()
        }
    }
}

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

    runner.register('Purging old dumps', settings.PURGE_OLD_DUMPS_INTERVAL, ctx =>
        purgeOldDumps(
            connection,
            dumpManager,
            uploadManager,
            settings.STORAGE_ROOT,
            settings.DBS_DIR_MAXIMUM_SIZE_BYTES,
            ctx
        )
    )

    runner.register('Cleaning failed uploads', settings.CLEAN_FAILED_UPLOADS_INTERVAL, cleanFailedUploads)

    runner.register('Updating metrics', settings.UPDATE_QUEUE_SIZE_GAUGE_INTERVAL, () =>
        updateQueueSizeGauge(uploadManager)
    )

    runner.run()
}
