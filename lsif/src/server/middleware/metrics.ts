import * as metrics from '../metrics'
import express from 'express'
import onFinished from 'on-finished'
import promClient from 'prom-client'

/**
 * Middleware function used to emit HTTP durations for LSIF functions. Originally
 * we used an express bundle, but that did not allow us to have different histogram
 * bucket for different endpoints, which makes half of the metrics useless in the
 * presence of large uploads.
 */
export const metricsMiddleware = (req: express.Request, res: express.Response, next: express.NextFunction): void => {
    let histogram: promClient.Histogram | undefined
    switch (req.path) {
        case '/upload':
            histogram = metrics.httpUploadDurationHistogram
            break

        case '/exists':
        case '/request':
            histogram = metrics.httpQueryDurationHistogram
    }

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
