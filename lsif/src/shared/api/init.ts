import express from 'express'
import { createMetaRouter } from './routes/meta'
import { default as tracingMiddleware } from 'express-opentracing'
import { errorHandler } from './middleware/errors'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { makeMetricsMiddleware } from './middleware/metrics'
import { Tracer } from 'opentracing'
import promClient from 'prom-client'

export function makeExpressApp({
    routes,
    logger,
    tracer,
    histogramSelector = () => undefined,
}: {
    routes: express.Router[]
    logger: Logger
    tracer?: Tracer
    histogramSelector?: (path: string) => promClient.Histogram<string> | undefined
}): express.Express {
    const loggingMiddlewareConfig = {
        winstonInstance: logger,
        level: 'debug',
        ignoredRoutes: ['/ping', '/healthz', '/metrics'],
        requestWhitelist: ['method', 'url'],
        msg: 'Handled request',
    }

    const app = express()
    app.use(tracingMiddleware({ tracer }))
    app.use(loggingMiddleware(loggingMiddlewareConfig))
    app.use(makeMetricsMiddleware(histogramSelector))

    for (const route of routes) {
        app.use(route)
    }

    app.use(createMetaRouter())
    app.use(errorHandler(logger)) // must be last
    return app
}
