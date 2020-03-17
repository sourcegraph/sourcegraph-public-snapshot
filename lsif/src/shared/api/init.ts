import express from 'express'
import promClient from 'prom-client'
import { default as tracingMiddleware } from 'express-opentracing'
import { errorHandler } from './middleware/errors'
import { logger as loggingMiddleware } from 'express-winston'
import { makeMetricsMiddleware } from './middleware/metrics'
import { Tracer } from 'opentracing'
import { Logger } from 'winston'

export function makeExpressApp({
    routes,
    logger,
    tracer,
    selectHistogram = () => undefined,
}: {
    routes: express.Router[]
    logger: Logger
    tracer?: Tracer
    selectHistogram?: (route: string) => promClient.Histogram<string> | undefined
}): express.Express {
    const loggingOptions = {
        winstonInstance: logger,
        level: 'debug',
        ignoredRoutes: ['/ping', '/healthz', '/metrics'],
        requestWhitelist: ['method', 'url'],
        msg: 'Handled request',
    }

    const app = express()
    app.use(tracingMiddleware({ tracer }))
    app.use(loggingMiddleware(loggingOptions))
    app.use(makeMetricsMiddleware(selectHistogram))

    for (const route of routes) {
        app.use(route)
    }

    // Error handler must be registered last so its exception handlers
    // will apply to all routes and other middleware.
    app.use(errorHandler(logger))

    return app
}
