import promClient from 'prom-client'

export async function instrument<T>(
    durationHistogram: promClient.Histogram,
    errorsCounter: promClient.Counter,
    fn: () => Promise<T>
): Promise<T> {
    const end = durationHistogram.startTimer()
    try {
        return await fn()
    } catch (e) {
        errorsCounter.inc()
        throw e
    } finally {
        end()
    }
}
