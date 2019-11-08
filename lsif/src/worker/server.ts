import express from 'express'
import promClient from 'prom-client'
import { Logger } from 'winston'
import { WORKER_METRICS_PORT } from './settings'

/**
 * Create an express server that only has /healthz and /metric endpoints.
 *
 * @param logger The logger instance.
 */
export function startMetricsServer(logger: Logger): void {
    const app = express()
    app.get('/healthz', (_, res) => res.send('ok'))
    app.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    app.listen(WORKER_METRICS_PORT, () => logger.debug('listening', { port: WORKER_METRICS_PORT }))
}
