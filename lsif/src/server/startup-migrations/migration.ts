import { migrateFilenames } from './filenames'
import { clearOldRedisData } from './redis'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'

/**
 * Run all startup migrations.
 *
 * @param ctx The tracing context.
 */
export function migrate(ctx: TracingContext): Promise<void> {
    return logAndTraceCall(ctx, 'Running startup migrations', async () => {
        await migrateFilenames(ctx)
        await clearOldRedisData(ctx)
    })
}
