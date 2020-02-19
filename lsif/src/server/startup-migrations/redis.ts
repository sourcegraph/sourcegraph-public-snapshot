import { createClient } from 'redis'
import { promisify } from 'util'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'
import { createSilentLogger } from '../../shared/logging'

/**
 * Remove all old LSIF data from redis.
 *
 * @param ctx The tracing context.
 */
export function clearOldRedisData(ctx: TracingContext): Promise<void> {
    return logAndTraceCall(ctx, 'Cleaning old redis data', async ({ logger = createSilentLogger() }) => {
        const script = `
            for i, key in ipairs(redis.call('keys', 'lsif:*')) do
                redis.call('del', key)
            end

            for i, key in ipairs(redis.call('keys', 'bull:*')) do
                redis.call('del', key)
            end
        `

        const url = process.env.REDIS_STORE_ENDPOINT || process.env.REDIS_ENDPOINT || 'redis-store:6379'
        const urlWithProtocol = url.includes('//') ? url : `redis://${url}`

        try {
            const client = createClient(urlWithProtocol)
            const evalAsync: (script: string, numArgs: number) => Promise<void> = promisify(client.eval).bind(client)
            await evalAsync(script, 0)
        } catch (err) {
            logger.warning('Failed to clean old LSIF data from redis-store', { error: err })
        }
    })
}
