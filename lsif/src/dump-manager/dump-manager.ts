import * as constants from '../shared/constants'
import * as path from 'path'
import * as settings from './settings'
import promClient from 'prom-client'
import { createLogger } from '../shared/logging'
import { ensureDirectory } from '../shared/paths'
import { Logger } from 'winston'
import express from 'express'

/**
 * No-op dump-manager process.
 *
 * @param logger The logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_' })

    // Ensure storage roots exist
    await ensureDirectory(settings.STORAGE_ROOT)
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(settings.STORAGE_ROOT, constants.UPLOADS_DIR))

    const app = express()
    app.get('/ping', (_, res) => res.send('ok'))
    app.get('/healthz', (_, res) => res.send('ok'))
    app.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    app.listen(settings.HTTP_PORT, () => logger.debug('LSIF dump manager listening on', { port: settings.HTTP_PORT }))
}

// Initialize logger
const appLogger = createLogger('lsif-dump-manager')

// Launch!
main(appLogger).catch(error => {
    appLogger.error('Failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
