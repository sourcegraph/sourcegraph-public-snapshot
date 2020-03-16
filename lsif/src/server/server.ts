import * as metrics from './metrics'
import * as settings from './settings'
import promClient from 'prom-client'
import { Backend } from './backend/backend'
import { createLogger } from '../shared/logging'
import { createLsifRouter } from './routes/lsif'
import { createPostgresConnection } from '../shared/database/postgres'
import { createTracer } from '../shared/tracing'
import { createUploadRouter } from './routes/uploads'
import { Logger } from 'winston'
import { startTasks } from './tasks/runner'
import { UploadManager } from '../shared/store/uploads'
import { waitForConfiguration } from '../shared/config/config'
import { DumpManager } from '../shared/store/dumps'
import { DependencyManager } from '../shared/store/dependencies'
import { SRC_FRONTEND_INTERNAL } from '../shared/config/settings'
import { makeExpressApp } from '../shared/api/init'
import { createInternalRouter } from './routes/internal'

/**
 * Runs the HTTP server that accepts LSIF dump uploads and responds to LSIF requests.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-server', fetchConfiguration())

    // Create database connection and entity wrapper classes
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const dumpManager = new DumpManager(connection)
    const uploadManager = new UploadManager(connection)
    const dependencyManager = new DependencyManager(connection)
    const backend = new Backend(dumpManager, dependencyManager, SRC_FRONTEND_INTERNAL)

    // Start background tasks
    startTasks(connection, dumpManager, uploadManager, logger)

    // Register middleware and serve
    const app = makeExpressApp({
        routes: [
            createUploadRouter(dumpManager, uploadManager, logger),
            createLsifRouter(backend, uploadManager, logger, tracer),
            createInternalRouter(dumpManager, uploadManager, logger),
        ],
        logger,
        tracer,
        histogramSelector,
    })

    app.listen(settings.HTTP_PORT, () => logger.debug('LSIF API server listening on', { port: settings.HTTP_PORT }))
}

function histogramSelector(route: string): promClient.Histogram<string> | undefined {
    switch (route) {
        case '/upload':
            return metrics.httpUploadDurationHistogram

        case '/exists':
        case '/request':
            return metrics.httpQueryDurationHistogram
    }

    return undefined
}

// Initialize logger
const appLogger = createLogger('lsif-server')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
