import { Connection } from 'typeorm'
import { DumpManager } from '../../shared/store/dumps'
import { migrateFilenames } from './filenames'
import { clearOldRedisData } from './redis'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'

/**
 * Run all startup migrations.
 *
 * @param ctx The tracing context.
 * @param connection The Postgres connection.
 * @param dumpManager The dump manager.
 */
export function migrate(ctx: TracingContext, connection: Connection, dumpManager: DumpManager): Promise<void> {
    return logAndTraceCall(ctx, 'Running startup migrations', async () => {
        await migrateFilenames(ctx)
        await clearOldRedisData(ctx)
    })
}
