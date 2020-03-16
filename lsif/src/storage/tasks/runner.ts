import * as settings from '../settings'
import { cleanFailedUploads, purgeOldDumps } from './uploads'
import { Connection } from 'typeorm'
import { Logger } from 'winston'
import { TaskRunner } from '../../shared/tasks'

/**
 * Begin running cleanup tasks on a schedule in the background.
 *
 * @param connection The Postgres connection.
 * @param logger The logger instance.
 */
export function startTasks(connection: Connection, logger: Logger): void {
    const runner = new TaskRunner(connection, logger)
    runner.register('Cleaning failed uploads', settings.CLEAN_FAILED_UPLOADS_INTERVAL, ctx => cleanFailedUploads(ctx))

    runner.register('Purging old dumps', settings.PURGE_OLD_DUMPS_INTERVAL, ctx =>
        purgeOldDumps(settings.STORAGE_ROOT, settings.DBS_DIR_MAXIMUM_SIZE_BYTES, ctx)
    )

    runner.run()
}
