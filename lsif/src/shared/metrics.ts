import promClient, { labelValues } from 'prom-client'

export async function instrument<T>(
    durationHistogram: promClient.Histogram,
    errorsCounter: promClient.Counter,
    fn: () => Promise<T>
): Promise<T> {
    return instrumentWithLabels(durationHistogram, errorsCounter, {}, fn)
}

export async function instrumentWithLabels<T>(
    durationHistogram: promClient.Histogram,
    errorsCounter: promClient.Counter,
    labels: labelValues,
    fn: () => Promise<T>
): Promise<T> {
    const end = durationHistogram.startTimer(labels)
    try {
        return await fn()
    } catch (error) {
        errorsCounter.inc()
        throw error
    } finally {
        end()
    }
}
