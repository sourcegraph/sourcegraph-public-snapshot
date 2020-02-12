import { migrateFilenames } from './filenames'
import { clearOldRedisData } from './redis'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'
import { assignIndexer } from './indexers'
import { Connection } from 'typeorm'

/**
 * Run all startup migrations.
 *
 * @param connection The Postgres connection.
 * @param ctx The tracing context.
 */
export function migrate(connection: Connection, ctx: TracingContext): Promise<void> {
    return logAndTraceCall(ctx, 'Running startup migrations', async () => {
        await migrateFilenames(ctx)
        await clearOldRedisData(ctx)
        await assignIndexer(connection, ctx)
    })
}
