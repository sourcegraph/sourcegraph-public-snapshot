import promClient from 'prom-client'

/**
 * TODO
 *
 * @param durationHistogram
 * @param errorsCounter
 * @param fn
 */
export async function instrument<T>(
    durationHistogram: promClient.Histogram,
    errorsCounter: promClient.Counter,
    fn: () => Promise<T>
): Promise<T> {
    const end = durationHistogram.startTimer()
    try {
        return await fn()
    } catch (error) {
        errorsCounter.inc()
        throw error
    } finally {
        end()
    }
}
