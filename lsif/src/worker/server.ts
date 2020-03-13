import * as settings from './settings'
import express from 'express'
import promClient from 'prom-client'
import { Logger } from 'winston'
import { logger as loggingMiddleware } from 'express-winston'

/**
 * Create an express server that only has /healthz and /metric endpoints.
 *
 * @param logger The logger instance.
 */
export function startMetricsServer(logger: Logger): void {
    const app = express()
    app.use(
        loggingMiddleware({
            winstonInstance: logger,
            level: 'debug',
            ignoredRoutes: ['/ping', '/healthz', '/metrics'],
            requestWhitelist: ['method', 'url'],
            msg: 'Handled request',
        })
    )
    app.use(createMetaRouter)

    app.listen(settings.WORKER_METRICS_PORT, () =>
        logger.debug('Worker metrics server listening', { port: settings.WORKER_METRICS_PORT })
    )
}

/** Create a router containing health and metrics endpoint. */
export function createMetaRouter(): express.Router {
    const router = express.Router()
    router.get('/ping', (_, res) => res.send('ok'))
    router.get('/healthz', (_, res) => res.send('ok'))
    router.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    return router
}
