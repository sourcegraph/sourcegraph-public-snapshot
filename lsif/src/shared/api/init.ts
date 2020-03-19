import express from 'express'
import promClient from 'prom-client'
import { default as tracingMiddleware } from 'express-opentracing'
import { errorHandler } from './middleware/errors'
import { logger as loggingMiddleware } from 'express-winston'
import { makeMetricsMiddleware } from './middleware/metrics'
import { Tracer } from 'opentracing'
import { Logger } from 'winston'

export function startExpressApp({
    routes,
    port,
    logger,
    tracer,
    selectHistogram = () => undefined,
}: {
    routes: express.Router[]
    port: number
    logger: Logger
    tracer?: Tracer
    selectHistogram?: (route: string) => promClient.Histogram<string> | undefined
}): void {
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
    app.use(createMetaRouter())

    for (const route of routes) {
        app.use(route)
    }

    // Error handler must be registered last so its exception handlers
    // will apply to all routes and other middleware.
    app.use(errorHandler(logger))

    app.listen(port, () => logger.debug('API server listening', { port }))
}

/** Create a router containing health and metrics endpoint. */
function createMetaRouter(): express.Router {
    const router = express.Router()
    router.get('/ping', (_, res) => res.send('ok'))
    router.get('/healthz', (_, res) => res.send('ok'))
    router.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    return router
}
