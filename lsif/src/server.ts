import * as definitionsSchema from './lsif.schema.json'
import * as fs from 'mz/fs'
import * as path from 'path'
import Ajv from 'ajv'
import bodyParser from 'body-parser'
import exitHook from 'async-exit-hook'
import express from 'express'
import onFinished from 'on-finished'
import onHeaders from 'on-headers'
import promClient from 'prom-client'
import uuid from 'uuid'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { createDatabaseFilename, createDirectory, hasErrorCode, readEnvInt } from './util'
import { Database } from './database'
import { Job, Queue, Scheduler } from 'node-resque'
import { logger, initLogger } from './logger'
import { RealQueue, rewriteJobMeta, WorkerMeta } from './queue'
import { validateLsifInput } from './input'
import { wrap } from 'async-middleware'
import { XrepoDatabase } from './xrepo'
import {
    CONNECTION_CACHE_CAPACITY_GAUGE,
    DOCUMENT_CACHE_CAPACITY_GAUGE,
    RESULT_CHUNK_CACHE_CAPACITY_GAUGE,
    QUEUE_SIZE_GAUGE,
    HTTP_QUERY_DURATION_HISTOGRAM,
    HTTP_UPLOAD_DURATION_HISTOGRAM,
} from './metrics'

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const HTTP_PORT = readEnvInt('HTTP_PORT', 3186)

/**
 * The host running the redis instance containing work queues. Defaults to localhost.
 */
const REDIS_HOST = process.env.REDIS_HOST || 'localhost'

/**
 * The port of the redis instance containing work queues. Defaults to 6379.
 */
const REDIS_PORT = readEnvInt('REDIS_PORT', 6379)

/**
 * The number of SQLite connections that can be opened at once. This
 * value may be exceeded for a short period if many handles are held
 * at once.
 */
const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 100)

/**
 * The maximum number of documents that can be held in memory at once.
 */
const DOCUMENT_CACHE_CAPACITY = readEnvInt('DOCUMENT_CACHE_CAPACITY', 1024 * 1024 * 1024)

/**
 * The maximum number of result chunks that can be held in memory at once.
 */
const RESULT_CHUNK_CACHE_CAPACITY = readEnvInt('RESULT_CHUNK_CACHE_CAPACITY', 1024 * 1024 * 1024)

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Whether or not to disable input validation. Validation is enabled by default.
 */
const DISABLE_VALIDATION = process.env.DISABLE_VALIDATION === 'true'

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    // Initialize logger
    initLogger('lsif-server')

    // Collect process metrics
    promClient.collectDefaultMetrics({ prefix: 'lsif_server_' })

    // Update cache capacities on startup
    CONNECTION_CACHE_CAPACITY_GAUGE.set(CONNECTION_CACHE_CAPACITY)
    DOCUMENT_CACHE_CAPACITY_GAUGE.set(DOCUMENT_CACHE_CAPACITY)
    RESULT_CHUNK_CACHE_CAPACITY_GAUGE.set(RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await createDirectory(STORAGE_ROOT)
    await createDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await createDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create queue to publish jobs for worker
    const queue = await setupQueue()

    // Create app + middleware
    const app = express()
    app.use(metricMiddleware)
    app.use(loggingMiddleware)

    // Register endpoints
    addMetaEndpoints(app)
    addQueueEndpoints(app, queue)
    addLsifEndpoints(app, queue)

    // Error handler must be registered last
    app.use(errorHandler)

    app.listen(HTTP_PORT, () => {
        logger.debug('Serving LSIF data', { port: HTTP_PORT })
    })
}

/**
 * Connect and start an active connection to the worker queue. We also run a
 * node-resque scheduler on each server instance, as these are guaranteed to
 * always be up with a responsive system. The schedulers will do their own
 * master election via a redis key and will check for dead workers attached
 * to the queue.
 */
async function setupQueue(): Promise<RealQueue> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
        namespace: 'lsif',
    }

    // Create and start queue
    const queue = new Queue({ connection: connectionOptions }) as RealQueue
    queue.on('error', e => logger.error('Queue error', e && e.message))
    await queue.connect()
    exitHook(() => queue.end())

    const emitQueueSizeMetric = (): void => {
        queue
            .queued('lsif', 0, -1)
            .then(
                jobs => QUEUE_SIZE_GAUGE.set(jobs.length),
                e => logger.error('Failed to get queued jobs', { error: e && e.message })
            )
    }

    // Update queue size metric on a timer
    setInterval(emitQueueSizeMetric, 1000)

    // Create scheduler and attach loggers to the interesting events
    const scheduler = new Scheduler({ connection: connectionOptions })
    scheduler.on('start', () => logger.debug('Started scheduler'))
    scheduler.on('end', () => logger.debug('Ended scheduler'))
    scheduler.on('poll', () => logger.debug('Checking for stuck workers'))
    scheduler.on('master', () => logger.debug('Scheduler has become master'))
    scheduler.on('cleanStuckWorker', (worker: string) => logger.debug('Cleaning stuck worker', { worker }))
    scheduler.on('transferredJob', (_: number, job: Job<any>) => logger.debug('Transferring job', { job }))
    scheduler.on('error', e => logger.error('Scheduler error', e && e.message))

    await scheduler.connect()
    exitHook(() => scheduler.end())
    scheduler.start().catch(e => logger.error('Failed to start scheduler', e && e.message))

    return queue
}

/**
 * Add common health and metrics endpoint.
 *
 * @param app The express app.
 */
function addMetaEndpoints(app: express.Application): void {
    app.get('/healthz', (req: express.Request, res: express.Response): void => {
        res.send('ok')
    })

    app.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })
}

/**
 * Add endpoints to the HTTP API to view/control the worker queue.
 *
 * @param app The express app.
 * @param queue The queue containing LSIF jobs.
 */
function addQueueEndpoints(app: express.Application, queue: RealQueue): void {
    app.get(
        '/queued',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const queuedJobs = await queue.queued('lsif', 0, -1)

                res.send(
                    queuedJobs.map(job => ({
                        ...rewriteJobMeta(job),
                    }))
                )
            }
        )
    )

    app.get(
        '/failed',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const failedJobs = await queue.failed(0, -1)
                failedJobs.sort((a, b) => a.failed_at.localeCompare(b.failed_at))

                res.send(
                    failedJobs.map(job => ({
                        error: job.error,
                        failed_at: new Date(job.failed_at).toISOString(),
                        ...rewriteJobMeta(job.payload),
                    }))
                )
            }
        )
    )

    app.get(
        '/active',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const workerMeta = Array.from(Object.values(await queue.allWorkingOn())).filter(
                    (x): x is WorkerMeta => x !== 'started'
                )
                workerMeta.sort((a, b) => a.run_at.localeCompare(b.run_at))

                res.send(
                    workerMeta.map(job => ({
                        started_at: new Date(job.run_at).toISOString(),
                        ...rewriteJobMeta(job.payload),
                    }))
                )
            }
        )
    )
}

/**
 * Add endpoints to the HTTP API to upload and query LSIF dumps.
 *
 * @param app The express app.
 * @param queue The queue containing LSIF jobs.
 */
function addLsifEndpoints(app: express.Application, queue: RealQueue): void {
    // Compile schema with defs as a reference
    const validator = new Ajv().addSchema({ $id: 'defs.json', ...definitionsSchema }).compile({
        oneOf: [{ $ref: 'defs.json#/definitions/Vertex' }, { $ref: 'defs.json#/definitions/Edge' }],
    })

    // Create cross-repos database
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_CAPACITY)
    const resultChunkCache = new ResultChunkCache(RESULT_CHUNK_CACHE_CAPACITY)
    const filename = path.join(STORAGE_ROOT, 'xrepo.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)

    // Factory function to open a database for a given repository/commit
    const createDatabase = async (repository: string, commit: string): Promise<Database | undefined> => {
        const file = createDatabaseFilename(STORAGE_ROOT, repository, commit)

        try {
            await fs.stat(file)
        } catch (e) {
            if (hasErrorCode(e, 'ENOENT')) {
                return undefined
            }

            throw e
        }

        return new Database(
            STORAGE_ROOT,
            xrepoDatabase,
            connectionCache,
            documentCache,
            resultChunkCache,
            repository,
            commit,
            file
        )
    }

    app.post(
        '/upload',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit } = req.query
                checkRepository(repository)
                checkCommit(commit)

                const filename = path.join(STORAGE_ROOT, 'uploads', uuid.v4())
                const output = fs.createWriteStream(filename)

                try {
                    await validateLsifInput(req, output, DISABLE_VALIDATION ? undefined : validator)
                } catch (e) {
                    throw Object.assign(e, { status: 422 })
                }

                // Enqueue input job
                logger.info('Enqueueing conversion job', { repository, commit })
                await queue.enqueue('lsif', 'convert', [repository, commit, filename])
                res.json(null)
            }
        )
    )

    app.post(
        '/exists',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, file } = req.query
                checkRepository(repository)
                checkCommit(commit)

                const db = await createDatabase(repository, commit)
                if (!db) {
                    res.json(false)
                    return
                }

                const result = !file || (await db.exists(file))
                res.json(result)
            }
        )
    )

    app.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit } = req.query
                const { path, position, method } = req.body
                checkRepository(repository)
                checkCommit(commit)
                checkMethod(method, ['definitions', 'references', 'hover'])
                const cleanMethod = method as 'definitions' | 'references' | 'hover'

                const db = await createDatabase(repository, commit)
                if (!db) {
                    throw Object.assign(new Error(`No LSIF data available for ${repository}@${commit}.`), {
                        status: 404,
                    })
                }

                res.json(await db[cleanMethod](path, position))
            }
        )
    )
}

/**
 * Middleware function used to emit HTTP durations for LSIF functions. Originally
 * we used an express bundle, but that did not allow us to have different histogram
 * bucket for different endpoints, which makes half of the metrics useless in the
 * presence of large uploads.
 */
function metricMiddleware(req: express.Request, res: express.Response, next: express.NextFunction): void {
    let histogram: promClient.Histogram | undefined

    switch (req.path) {
        case '/upload':
            histogram = HTTP_UPLOAD_DURATION_HISTOGRAM
            break

        case '/exists':
        case '/request':
            histogram = HTTP_QUERY_DURATION_HISTOGRAM
    }

    if (histogram !== undefined) {
        const labels = { code: 0 }
        const end = histogram.startTimer(labels)

        onFinished(res, () => {
            labels.code = res.statusCode
            end()
        })
    }

    next()
}

/**
 * A pair of seconds and nanoseconds representing the output of
 * the nodejs high-resolution timer.
 */
type HrTime = [number, number]

/**
 * Middleware function used to log requests and the corresponding
 * response status code and wall time taken to process the request
 * (to the point where headers are emitted).
 */
function loggingMiddleware(req: express.Request, res: express.Response, next: express.NextFunction): void {
    const start = process.hrtime()
    let end: HrTime | undefined

    onHeaders(res, () => {
        end = process.hrtime()
    })

    onFinished(res, () => {
        const responseTime = end ? `${((end[0] - start[0]) * 1e3 + (end[1] - start[1]) * 1e-6).toFixed(3)}ms` : ''

        logger.debug('request', {
            method: req.method,
            path: req.path,
            statusCode: res.statusCode,
            responseTime,
        })
    })

    next()
}

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 */
function errorHandler(e: any, req: express.Request, res: express.Response, next: express.NextFunction): void {
    if (res.headersSent) {
        return next(e)
    }

    if (e && e.status) {
        res.status(e.status).send({ message: e.message })
        return
    }

    logger.error('Uncaught exception', { error: e && e.message })
    res.status(500).send({ message: 'Unknown error' })
}

/**
 * Throws an error with status 400 if the repository string is invalid.
 */
export function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error('Must specify the repository (usually of the form github.com/user/repo)'), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit string is invalid.
 */
export function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error('Must specify the commit as a 40 character hash ' + commit), { status: 400 })
    }
}

/**
 * Throws an error with status 422 if the requested method is not supported.
 */
export function checkMethod(method: string, supportedMethods: string[]): void {
    if (!supportedMethods.includes(method)) {
        throw Object.assign(new Error(`Method must be one of ${Array.from(supportedMethods).join(', ')}`), {
            status: 422,
        })
    }
}

main().catch(e => logger.error('Failed to start process', e && e.message))
