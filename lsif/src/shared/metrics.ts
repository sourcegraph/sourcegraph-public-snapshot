import promClient from 'prom-client'

/**
 * Instrument the duration and error rate of the given function.
 *
 * @param durationHistogram The histogram for operation durations.
 * @param errorsCounter The counter for errors.
 * @param fn The function to instrument.
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
