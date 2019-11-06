import { Queue } from 'bull'
import * as fs from 'mz/fs'
import * as path from 'path'
import bodyParser from 'body-parser'
import express from 'express'
import onFinished from 'on-finished'
import promClient from 'prom-client'
import uuid from 'uuid'
import { httpUploadDurationHistogram, httpQueryDurationHistogram, queueSizeGauge } from './server.metrics'
import { chunk } from 'lodash'
import cors from 'cors'
import {
    connectionCacheCapacityGauge,
    documentCacheCapacityGauge,
    resultChunkCacheCapacityGauge,
} from './cache.metrics'
import { dbFilename, dbFilenameOld, ensureDirectory, readEnvInt } from './util'
import { createPostgresConnection } from './connection'
import { Backend, ReferencePaginationCursor } from './backend'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'
import { wrap } from 'async-middleware'
import { XrepoDatabase } from './xrepo'
import { createTracer, logAndTraceCall, TracingContext, addTags } from './tracing'
import { Span, Tracer } from 'opentracing'
import { default as tracingMiddleware } from 'express-opentracing'
import { waitForConfiguration } from './config'
import { createLogger } from './logging'
import {
    enqueue,
    createQueue,
    ensureOnlyRepeatableJob,
    queueTypes,
    QUEUE_PREFIX,
    statesByQueue,
    ApiJobState,
} from './queue'
import { Connection } from 'typeorm'
import { LsifDump } from './xrepo.models'
import * as constants from './constants'
import pTimeout from 'p-timeout'
import { formatJob, formatJobFromMap } from './api-job'
import { Redis } from 'ioredis'
import * as settings from './settings'

const pipeline = promisify(_pipeline)

/**
 * Which port to run the LSIF server on. Defaults to 3186.
 */
const HTTP_PORT = readEnvInt('HTTP_PORT', 3186)

/**
 * The host and port running the redis instance containing work queues.
 *
 * Set addresses. Prefer in this order:
 *   - Specific envvar REDIS_STORE_ENDPOINT
 *   - Fallback envvar REDIS_ENDPOINT
 *   - redis-store:6379
 *
 *  Additionally keep this logic in sync with pkg/redispool/redispool.go and cmd/server/redis.go
 */
const REDIS_ENDPOINT = process.env.REDIS_STORE_ENDPOINT || process.env.REDIS_ENDPOINT || 'redis-store:6379'

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * The interval (in seconds) to schedule the update-tips job.
 */
const UPDATE_TIPS_JOB_SCHEDULE_INTERVAL = readEnvInt('UPDATE_TIPS_JOB_SCHEDULE_INTERVAL', 60 * 5)

/**
 * The interval (in seconds) to schedule the clean-old-jobs job.
 */
const CLEAN_OLD_JOBS_INTERVAL = readEnvInt('CLEAN_OLD_JOBS_INTERVAL', 60 * 60 * 8)

/**
 * The default number of remote dumps to open when performing a global find-reference operation.
 */
const DEFAULT_REFERENCES_NUM_REMOTE_DUMPS = 10

/**
 * The interval (in seconds) to schedule the clean-failed-jobs job.
 */
const CLEAN_FAILED_JOBS_INTERVAL = readEnvInt('CLEAN_FAILED_JOBS_INTERVAL', 60 * 60 * 8)

/**
 * The default page size for the job endpoints.
 */
const DEFAULT_JOB_PAGE_SIZE = readEnvInt('DEFAULT_JOB_PAGE_SIZE', 50)

/**
 * The maximum number of jobs to search in one call to the search-jobs.lua script.
 */
export const MAX_JOB_SEARCH = readEnvInt('MAX_JOB_SEARCH', 10000)

/**
 * The default page size for the dumps endpoint.
 */
const DEFAULT_DUMP_PAGE_SIZE = 50

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 *
 * @param logger The logger instance.
 */
const errorHandler = (
    logger: Logger
): ((error: any, req: express.Request, res: express.Response, next: express.NextFunction) => void) => (
    error: any,
    req: express.Request,
    res: express.Response,
    next: express.NextFunction
): void => {
    if (!error || !error.status) {
        // Only log errors that don't have a status attached
        logger.error('uncaught exception', { error })
    }

    if (!res.headersSent) {
        res.status((error && error.status) || 500).send({ message: (error && error.message) || 'Unknown error' })
    }
}

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
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

    // Update cache capacities on startup
    connectionCacheCapacityGauge.set(settings.CONNECTION_CACHE_CAPACITY)
    documentCacheCapacityGauge.set(settings.DOCUMENT_CACHE_CAPACITY)
    resultChunkCacheCapacityGauge.set(settings.RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, constants.DBS_DIR))
    await ensureDirectory(path.join(STORAGE_ROOT, constants.TEMP_DIR))
    await ensureDirectory(path.join(STORAGE_ROOT, constants.UPLOADS_DIR))

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration(), logger)
    const xrepoDatabase = new XrepoDatabase(STORAGE_ROOT, connection)
    const backend = new Backend(STORAGE_ROOT, xrepoDatabase, fetchConfiguration)

    // Temporary migrations
    await moveDatabaseFilesToSubdir() // TODO - remove after 3.12
    await ensureFilenamesAreIDs(connection) // TODO - remove after 3.10

    // Create queue to publish convert
    const queue = createQueue(REDIS_ENDPOINT, logger)

    // Schedule jobs on timers
    await ensureOnlyRepeatableJob(queue, 'update-tips', {}, UPDATE_TIPS_JOB_SCHEDULE_INTERVAL * 1000)
    await ensureOnlyRepeatableJob(queue, 'clean-old-jobs', {}, CLEAN_OLD_JOBS_INTERVAL * 1000)
    await ensureOnlyRepeatableJob(queue, 'clean-failed-jobs', {}, CLEAN_FAILED_JOBS_INTERVAL * 1000)

    // Update queue size metric on a timer
    setInterval(
        () =>
            queue
                .count()
                .then(count => queueSizeGauge.set(count))
                .catch(() => {}),
        1000
    )

    // Register the required commands on the queue's Redis client
    const scriptedClient = await defineRedisCommands(queue.client)

    const app = express()
    app.use(cors())

    if (tracer !== undefined) {
        app.use(tracingMiddleware({ tracer }))
    }

    app.use(
        loggingMiddleware({
            winstonInstance: logger,
            level: 'debug',
            ignoredRoutes: ['/ping', '/healthz', '/metrics'],
            requestWhitelist: ['method', 'url', 'query'],
            msg: 'request',
        })
    )
    app.use(metricsMiddleware)

    // Register endpoints
    app.use(metaEndpoints())
    app.use(lsifEndpoints(backend, queue, logger, tracer))
    app.use(dumpEndpoints(backend, logger, tracer))
    app.use(jobEndpoints(queue, scriptedClient, logger, tracer))

    // Error handler must be registered last
    app.use(errorHandler(logger))

    app.listen(HTTP_PORT, () => logger.debug('listening', { port: HTTP_PORT }))
}

/**
 * Move all db files in storage root to a subdirectory.
 */
async function moveDatabaseFilesToSubdir(): Promise<void> {
    for (const filename of await fs.readdir(STORAGE_ROOT)) {
        if (filename.endsWith('.db')) {
            await fs.rename(path.join(STORAGE_ROOT, filename), path.join(STORAGE_ROOT, constants.DBS_DIR, filename))
        }
    }
}

/**
 * If it hasn't been done already, migrate from the old pre-3.9 filename format
 * `$REPO@$COMMIT.lsif.db` to the new format `$ID.lsif.db`.
 */
async function ensureFilenamesAreIDs(db: Connection): Promise<void> {
    const doneFile = path.join(STORAGE_ROOT, 'id-based-filenames')
    if (await fs.exists(doneFile)) {
        // Already migrated.
        return
    }

    for (const dump of await db.getRepository(LsifDump).find()) {
        const oldFile = dbFilenameOld(STORAGE_ROOT, dump.repository, dump.commit)
        const newFile = dbFilename(STORAGE_ROOT, dump.id, dump.repository, dump.commit)
        if (!(await fs.exists(oldFile))) {
            continue
        }
        await fs.rename(oldFile, newFile)
    }

    // Create an empty done file to record that all files have been renamed.
    await fs.close(await fs.open(doneFile, 'w'))
}

/**
 * Middleware function used to emit HTTP durations for LSIF functions. Originally
 * we used an express bundle, but that did not allow us to have different histogram
 * bucket for different endpoints, which makes half of the metrics useless in the
 * presence of large uploads.
 */
function metricsMiddleware(req: express.Request, res: express.Response, next: express.NextFunction): void {
    let histogram: promClient.Histogram | undefined
    switch (req.path) {
        case '/upload':
            histogram = httpUploadDurationHistogram
            break

        case '/exists':
        case '/request':
            histogram = httpQueryDurationHistogram
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
 * Create a router containing health endpoint.
 */
function metaEndpoints(): express.Router {
    const router = express.Router()
    router.get('/ping', (_, res) => res.send('ok'))
    router.get('/healthz', (_, res) => res.send('ok'))
    router.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    return router
}

/**
 * Create a router containing the LSIF upload and query endpoints.
 *
 * @param backend The backend instance.
 * @param queue The queue containing LSIF jobs.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
function lsifEndpoints(backend: Backend, queue: Queue, logger: Logger, tracer: Tracer | undefined): express.Router {
    const router = express.Router()

    router.post(
        '/upload',
        wrap(
            async (req: express.Request & { span?: Span }, res: express.Response): Promise<void> => {
                const { repository, commit, root: rootRaw, blocking, maxWait } = req.query
                const root = normalizeRoot(rootRaw)
                const timeout = parseInt(maxWait, 10) || 0
                checkRepository(repository)
                checkCommit(commit)

                const ctx = createTracingContext(logger, req, {
                    repository,
                    commit,
                    root,
                })
                const filename = path.join(STORAGE_ROOT, constants.UPLOADS_DIR, uuid.v4())
                const output = fs.createWriteStream(filename)

                try {
                    await logAndTraceCall(ctx, 'uploading dump', () => pipeline(req, output))
                } catch (e) {
                    throw Object.assign(e, { status: 422 })
                }

                // Enqueue convert job
                logger.debug('enqueueing convert job', { repository, commit, root })
                const args = { repository, commit, root, filename }
                const job = await enqueue(queue, 'convert', args, {}, tracer, ctx.span)

                if (blocking) {
                    let promise = job.finished()
                    if (timeout >= 0) {
                        promise = pTimeout(promise, timeout * 1000)
                    }

                    try {
                        await promise

                        // Job succeeded while blocked, send success
                        res.status(200)
                        res.send({ id: job.id })
                        return
                    } catch (error) {
                        // Throw the job error, but not the timeout error
                        if (!error.message.includes('Promise timed out')) {
                            throw error
                        }
                    }
                }

                // Job will complete asynchronously, send a 202: Accepted with
                // the job id so that the client can continue to track the progress
                // asynchronously.

                res.status(202)
                res.send({ id: job.id })
            }
        )
    )

    router.post(
        '/exists',
        bodyParser.json({ limit: '1mb' }),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, file } = req.query
                checkRepository(repository)
                checkCommit(commit)
                checkFile(file)

                const ctx = createTracingContext(logger, req, { repository, commit })
                res.json(await backend.exists(repository, commit, file, ctx))
            }
        )
    )

    router.post(
        '/request',
        bodyParser.json({ limit: '1mb' }),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit } = req.query
                const { path: filePath, position, method } = req.body
                checkRepository(repository)
                checkCommit(commit)
                checkMethod(method, ['definitions', 'references', 'hover'])

                const ctx = createTracingContext(logger, req, { repository, commit })

                switch (method) {
                    case 'definitions':
                        res.json(await backend.definitions(repository, commit, filePath, position, ctx))
                        break

                    case 'hover':
                        res.json(await backend.hover(repository, commit, filePath, position, ctx))
                        break

                    case 'references': {
                        const { cursor: cursorRaw } = req.query
                        const limit = extractLimit(req, DEFAULT_REFERENCES_NUM_REMOTE_DUMPS)
                        const startCursor = parseCursor<ReferencePaginationCursor>(cursorRaw)

                        const { locations, cursor: endCursor } = await backend.references(
                            repository,
                            commit,
                            filePath,
                            position,
                            { limit, cursor: startCursor },
                            ctx
                        )

                        const encodedCursor = encodeCursor<ReferencePaginationCursor>(endCursor)
                        if (!encodedCursor) {
                            res.set('Link', nextLink(req, { limit, cursor: encodedCursor }))
                        }

                        res.json(locations)
                        break
                    }
                }
            }
        )
    )

    return router
}

/**
 * Create a router containing the LSIF dump endpoints.
 *
 * @param backend The backend instance.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
function dumpEndpoints(backend: Backend, logger: Logger, tracer: Tracer | undefined): express.Router {
    const router = express.Router()

    router.get(
        '/dumps/:repository',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository } = req.params
                const { query, visibleAtTip: visibleAtTipRaw } = req.query
                const { limit, offset } = limitOffset(req, DEFAULT_DUMP_PAGE_SIZE)
                const visibleAtTip = visibleAtTipRaw === 'true'
                const { dumps, totalCount } = await backend.dumps(
                    decodeURIComponent(repository),
                    query,
                    visibleAtTip,
                    limit,
                    offset
                )

                if (offset + dumps.length < totalCount) {
                    res.set('Link', nextLink(req, { limit, offset: offset + dumps.length }))
                }

                res.json({ dumps, totalCount })
            }
        )
    )

    router.get(
        '/dumps/:repository/:id',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, id } = req.params
                const dump = await backend.dump(parseInt(id, 10))
                if (!dump || dump.repository !== decodeURIComponent(repository)) {
                    throw Object.assign(new Error('LSIF dump not found'), { status: 404 })
                }

                res.json(dump)
            }
        )
    )

    return router
}

/**
 * Create a router containing the job endpoints.
 *
 * @param queue The queue instance.
 * @param scriptedClient The Redis client with scripts loaded.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
function jobEndpoints(
    queue: Queue,
    scriptedClient: ScriptedRedis,
    logger: Logger,
    tracer: Tracer | undefined
): express.Router {
    const router = express.Router()

    router.get(
        '/jobs/stats',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const counts = await queue.getJobCounts()

                res.send({
                    processingCount: counts.active,
                    erroredCount: counts.failed,
                    completedCount: counts.completed,
                    queuedCount: counts.waiting,
                    scheduledCount: counts.delayed,
                })
            }
        )
    )

    router.get(
        `/jobs/:state(${Array.from(queueTypes.keys()).join('|')})`,
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { state } = req.params as { state: ApiJobState }
                const { query } = req.query
                const { limit, offset } = limitOffset(req, DEFAULT_JOB_PAGE_SIZE)

                const queueName = queueTypes.get(state)
                if (!queueName) {
                    throw new Error(`Unknown job state ${state}`)
                }

                if (!query) {
                    const rawJobs = await queue.getJobs([queueName], offset, offset + limit - 1)
                    const jobs = rawJobs.map(job => formatJob(job, state))
                    const totalCount = (await queue.getJobCountByTypes([queueName])) as never

                    if (offset + jobs.length < totalCount) {
                        res.set('Link', nextLink(req, { limit, offset: offset + jobs.length }))
                    }

                    res.send({ jobs, totalCount })
                } else {
                    const [payloads, nextOffset] = await scriptedClient.searchJobs([
                        QUEUE_PREFIX,
                        queueName,
                        query,
                        offset,
                        limit,
                        MAX_JOB_SEARCH,
                    ])

                    const jobs = payloads
                        // Convert each hgetall response into a map
                        .map(payload => new Map(chunk(payload, 2) as [string, string][]))
                        // Format each job
                        .map(payload => formatJobFromMap(payload, state))

                    if (nextOffset) {
                        res.set('Link', nextLink(req, { limit, offset: nextOffset }))
                    }

                    res.send({ jobs })
                }
            }
        )
    )

    router.get(
        '/jobs/:id',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const job = await queue.getJob(req.params.id)
                if (!job) {
                    throw Object.assign(new Error('Job not found'), {
                        status: 404,
                    })
                }

                const rawState = await job.getState()
                const state = statesByQueue.get(rawState === 'waiting' ? 'wait' : rawState)
                if (!state) {
                    throw new Error(`Unknown job state ${state}.`)
                }

                res.send(formatJob(job, state))
            }
        )
    )

    return router
}

/**
 * The type of the Redis client with additional script commands defined.
 */
type ScriptedRedis = Redis & {
    // runs ./search-jobs.lua
    searchJobs: (args: (string | number | boolean)[]) => Promise<[string[][], number | null]>
}

/**
 * Registers the search-jobs.lua script in the given Redis instance. This function
 * returns the same redis client with additional methods attached.
 *
 * @param client The redis client.
 */
async function defineRedisCommands(client: Redis): Promise<ScriptedRedis> {
    client.defineCommand('searchJobs', {
        numberOfKeys: 2,
        lua: (await fs.readFile(`${__dirname}/search-jobs.lua`)).toString(),
    })

    // The defineCommand method on the client dynamically defines a new method, but
    // the type system doesn't know that. We need to do a dumb cast here. This only
    // requires us to know the return type of the script.
    return client as ScriptedRedis
}

/**
 * Throws an error with status 400 if the repository string is invalid.
 */
function checkRepository(repository: any): void {
    if (typeof repository !== 'string') {
        throw Object.assign(new Error(`Must specify a repository ${repository}`), {
            status: 400,
        })
    }
}

/**
 * Throws an error with status 400 if the commit string is invalid.
 */
function checkCommit(commit: any): void {
    if (typeof commit !== 'string' || commit.length !== 40 || !/^[0-9a-f]+$/.test(commit)) {
        throw Object.assign(new Error(`Must specify the commit as a 40 character hash ${commit}`), { status: 400 })
    }
}

/**
 * Throws an error with status 400 if the file is not present.
 */
function checkFile(file: any): void {
    if (typeof file !== 'string') {
        throw Object.assign(new Error(`Must specify a file ${file}`), { status: 400 })
    }
}

/**
 * Throws an error with status 422 if the requested method is not supported.
 */
function checkMethod(method: string, supportedMethods: string[]): void {
    if (!supportedMethods.includes(method)) {
        throw Object.assign(new Error(`Method must be one of ${Array.from(supportedMethods).join(', ')}`), {
            status: 422,
        })
    }
}

/**
 * Adds a trailing slash to a root unless it refers to the top level.
 *
 * - 'foo' -> 'foo/'
 * - 'foo/' -> 'foo/'
 * - '/' -> ''
 * - '' -> ''
 */
function normalizeRoot(root: any): string {
    if (root === undefined || root === '/' || root === '') {
        return ''
    }
    if (typeof root !== 'string') {
        throw Object.assign(new Error('root must be a string'), {
            status: 422,
        })
    }
    return root.endsWith('/') ? root : root + '/'
}

/**
 * Create a tracing context from the request logger and tracing span
 * tagged with the given values.
 *
 * @param logger The logger instance.
 * @param req The express request.
 * @param tags The tags to apply to the logger and span.
 */
function createTracingContext(
    logger: Logger,
    req: express.Request & { span?: Span },
    tags: { [K: string]: any }
): TracingContext {
    return addTags({ logger, span: req.span }, tags)
}

/**
 * Extract a page limit from the request query string.
 *
 * @param req The HTTP request.
 * @param defaultLimit The limit to use if one is not supplied.
 */
function extractLimit(req: express.Request, defaultLimit: number): number {
    return parseInt(req.query.limit, 10) || defaultLimit
}

/**
 * Extract a page limit and offset from the request query string.
 *
 * @param req The HTTP request.
 * @param defaultLimit The limit to use if one is not supplied.
 */
function limitOffset(req: express.Request, defaultLimit: number): { limit: number; offset: number } {
    return {
        limit: parseInt(req.query.limit, 10) || defaultLimit,
        offset: parseInt(req.query.offset, 10) || 0,
    }
}

/**
 * Create a link header payload with a next link based on the previous endpoint.
 *
 * @param req The HTTP request.
 * @param params The query params to overwrite.
 */
function nextLink(req: express.Request, params: { [K: string]: any }): string {
    const url = new URL(`${req.protocol}://${req.get('host')}${req.originalUrl}`)
    for (const [key, value] of Object.entries(params)) {
        url.searchParams.set(key, String(value))
    }

    return `<${url.href}>; rel="next"`
}

/**
 * Parse a base64-encoded JSON payload into the expected type.
 *
 * @param cursorRaw The raw cursor.
 */
function parseCursor<T>(cursorRaw: any): T | undefined {
    if (cursorRaw === undefined) {
        return undefined
    }

    try {
        return JSON.parse(new Buffer(cursorRaw, 'base64').toString('ascii'))
    } catch {
        throw Object.assign(new Error(`Malformed cursor supplied ${cursorRaw}`), { status: 400 })
    }
}

function encodeCursor<T>(cursor: T | undefined): string | undefined {
    if (!cursor) {
        return undefined
    }

    return new Buffer(JSON.stringify(cursor)).toString('base64')
}

// Initialize logger
const appLogger = createLogger('lsif-server')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
