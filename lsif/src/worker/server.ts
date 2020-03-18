import * as settings from './settings'
import { Logger } from 'winston'
import { makeExpressApp } from '../shared/api/init'

/** Create an express server containing health and metrics endpoint. */
export function startMetricsServer(logger: Logger): void {
    const app = makeExpressApp({
        routes: [],
        logger,
    })

    app.listen(settings.WORKER_METRICS_PORT, () =>
        logger.debug('Worker metrics server listening', { port: settings.WORKER_METRICS_PORT })
    )
}
