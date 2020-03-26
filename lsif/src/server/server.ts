import * as metrics from './metrics'
import * as settings from './settings'
import promClient from 'prom-client'
import { Backend } from './backend/backend'
import { createLogger } from '../shared/logging'
import { createLsifRouter } from './routes/lsif'
import { createPostgresConnection } from '../shared/database/postgres'
import { createTracer } from '../shared/tracing'
import { createUploadRouter } from './routes/uploads'
import { ensureDirectory } from '../shared/paths'
import { Logger } from 'winston'
import { startTasks } from './tasks'
import { UploadManager } from '../shared/store/uploads'
import { waitForConfiguration } from '../shared/config/config'
import { DumpManager } from '../shared/store/dumps'
import { DependencyManager } from '../shared/store/dependencies'
import { SRC_FRONTEND_INTERNAL } from '../shared/config/settings'
import { startExpressApp } from '../shared/api/init'
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

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)

    // Create database connection and entity wrapper classes
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const dumpManager = new DumpManager(connection)
    const uploadManager = new UploadManager(connection)
    const dependencyManager = new DependencyManager(connection)
    const backend = new Backend(dumpManager, dependencyManager, SRC_FRONTEND_INTERNAL)

    // Start background tasks
    startTasks(connection, uploadManager, logger)

    const routers = [
        createUploadRouter(dumpManager, uploadManager, logger),
        createLsifRouter(connection, backend, uploadManager, logger, tracer),
        createInternalRouter(dumpManager, uploadManager, logger),
    ]

    // Start server
    startExpressApp({ port: settings.HTTP_PORT, routers, logger, tracer, selectHistogram })
}

function selectHistogram(route: string): promClient.Histogram<string> | undefined {
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
