import * as settings from './settings'
import express from 'express'
import promClient from 'prom-client'
import { Logger } from 'winston'

/** Create an express server containing health and metrics endpoint. */
export function startMetricsServer(logger: Logger): void {
    const app = express()
    app.get('/ping', (_, res) => res.send('ok'))
    app.get('/healthz', (_, res) => res.send('ok'))
    app.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    app.listen(settings.WORKER_METRICS_PORT, () =>
        logger.debug('Worker metrics server listening', { port: settings.WORKER_METRICS_PORT })
    )
}
