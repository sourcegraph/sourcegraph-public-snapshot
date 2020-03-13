import express from 'express'
import onFinished from 'on-finished'
import promClient from 'prom-client'

/**
 * Middleware function used to emit HTTP durations for LSIF functions. Originally
 * we used an express bundle, but that did not allow us to have different histogram
 * bucket for different endpoints, which makes half of the metrics useless in the
 * presence of large uploads.
 */
export const makeMetricsMiddleware = (
    histogramSelector: (path: string) => promClient.Histogram<string> | undefined
) => <T>(req: express.Request, res: express.Response<T>, next: express.NextFunction): void => {
    const histogram = histogramSelector(req.path)

    if (histogram !== undefined) {
        const labels = { code: 0 }
        const end = histogram.startTimer(labels)

        onFinished(res, () => {
            labels.code = res.statusCode
            end()
        })
    }

    next()
}
