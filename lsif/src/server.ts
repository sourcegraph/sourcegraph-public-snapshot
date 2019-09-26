import * as fs from 'mz/fs'
import * as path from 'path'
import bodyParser from 'body-parser'
import exitHook from 'async-exit-hook'
import express from 'express'
import promBundle from 'express-prom-bundle'
import uuid from 'uuid'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { connectionCacheCapacityGauge, documentCacheCapacityGauge, resultChunkCacheCapacityGauge } from './metrics'
import { createDatabaseFilename, ensureDirectory, hasErrorCode, readEnvInt } from './util'
import { createGzip } from 'zlib'
import { createLogger } from './logger'
import { createPostgresConnection } from './connection'
import { Database } from './database.js'
import { Edge, Vertex } from 'lsif-protocol'
import { GITSERVER_URLS, updateCommits } from './commits'
import { identity } from 'lodash'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { pipeline as _pipeline, Readable } from 'stream'
import { promisify } from 'util'
import { Queue, Scheduler } from 'node-resque'
import { readGzippedJsonElements, stringifyJsonLines, validateLsifElements } from './input'
import { wrap } from 'async-middleware'
import { XrepoDatabase } from './xrepo'

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
 * The JSON schema validation function to use. If validation is disabled, then
 * this method has no observable behavior.
 */
const validateIfEnabled: (data: AsyncIterable<unknown>) => AsyncIterable<Vertex | Edge> = DISABLE_VALIDATION
    ? identity
    : validateLsifElements

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 *
 * @param logger The server's logger instance.
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
 * @param logger The application logger instance.
 */
async function main(logger: Logger): Promise<void> {
    // Update cache capacities on startup
    connectionCacheCapacityGauge.set(CONNECTION_CACHE_CAPACITY)
    documentCacheCapacityGauge.set(DOCUMENT_CACHE_CAPACITY)
    resultChunkCacheCapacityGauge.set(RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await ensureDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create queue to publish jobs for worker
    const queue = await setupQueue(logger)

    const app = express()
    app.use(
        loggingMiddleware({
            winstonInstance: logger,
            level: 'debug',
            ignoredRoutes: ['/ping', '/healthz', '/metrics'],
            requestWhitelist: ['method', 'url', 'query'],
            msg: 'request',
        })
    )
    app.use(promBundle({}))

    // Register endpoints
    app.use(metaEndpoints())
    app.use(await lsifEndpoints(queue, logger))

    // Error handler must be registered last
    app.use(errorHandler(logger))

    app.listen(HTTP_PORT, () => logger.debug('listening', { port: HTTP_PORT }))
}

/**
 * Connect and start an active connection to the worker queue. We also run a
 * node-resque scheduler on each server instance, as these are guaranteed to
 * always be up with a responsive system. The schedulers will do their own
 * master election via a redis key and will check for dead workers attached
 * to the queue.
 *
 * @param logger The server's logger instance.
 */
async function setupQueue(logger: Logger): Promise<Queue> {
    const [host, port] = REDIS_ENDPOINT.split(':', 2)

    const connectionOptions = {
        host,
        port: parseInt(port, 10),
        namespace: 'lsif',
    }

    // Create queue and log the interesting events
    const queue = new Queue({ connection: connectionOptions })
    queue.on('error', error => logger.error('queue error', { error }))
    await queue.connect()
    exitHook(() => queue.end())

    // Create scheduler log the interesting events
    const scheduler = new Scheduler({ connection: connectionOptions })
    scheduler.on('start', () => logger.debug('scheduler started'))
    scheduler.on('end', () => logger.debug('scheduler ended'))
    scheduler.on('poll', () => logger.debug('scheduler checking for stuck workers'))
    scheduler.on('master', () => logger.debug('scheduler became master'))
    scheduler.on('cleanStuckWorker', worker => logger.debug('scheduler cleaning stuck worker', { worker }))
    scheduler.on('transferredJob', (_, job) => logger.debug('scheduler transferring job', { job }))
    scheduler.on('error', error => logger.error('scheduler error', { error }))

    await scheduler.connect()
    exitHook(() => scheduler.end())
    await scheduler.start()

    return queue
}

/**
 * Create a router containing health endpoint.
 */
function metaEndpoints(): express.Router {
    const router = express.Router()
    router.get('/ping', (_, res) => res.send('ok'))
    router.get('/healthz', (_, res) => res.send('ok'))
    return router
}

/**
 * Create a router containing the LSIF upload and query endpoints.
 *
 * @param queue The queue containing LSIF jobs.
 * @param logger The server's logger instance.
 */
async function lsifEndpoints(queue: Queue, logger: Logger): Promise<express.Router> {
    const router = express.Router()

    // Create cross-repo database
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_CAPACITY)
    const resultChunkCache = new ResultChunkCache(RESULT_CHUNK_CACHE_CAPACITY)

    // Create cross-repo database
    const connection = await createPostgresConnection()
    const xrepoDatabase = new XrepoDatabase(connection)

    const createDatabaseFilenameStat = async (repository: string, commit: string): Promise<string | undefined> => {
        const file = createDatabaseFilename(STORAGE_ROOT, repository, commit)

        try {
            await fs.stat(file)
        } catch (e) {
            if (hasErrorCode(e, 'ENOENT')) {
                return undefined
            }

            throw e
        }

        return file
    }

    const tryLoadDatabase = async (repository: string, commit: string): Promise<Database | undefined> => {
        const file = await createDatabaseFilenameStat(repository, commit)
        if (!file) {
            return undefined
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

    const loadDatabase = async (repository: string, commit: string): Promise<Database | undefined> => {
        // Try to construct database for the exact commit
        const database = await tryLoadDatabase(repository, commit)
        if (database) {
            return database
        }

        // TODO - log here as well

        // If we don't know about this commit, request updated commit data
        // from the gitserver. This will pull back ancestors for this commit
        // up to a certain (configurable) depth and insert them into the
        // cross-repository database. This populates the necessary data for
        // the following query.
        if (!(await xrepoDatabase.isCommitTracked(repository, commit))) {
            await updateCommits(GITSERVER_URLS, xrepoDatabase, repository, commit, logger)
        }

        // Determine the closest commit that we actually have LSIF data for
        const commitWithData = await xrepoDatabase.findClosestCommitWithData(repository, commit)
        if (!commitWithData) {
            return undefined
        }

        // Try to construct a database for the approximate commit
        return tryLoadDatabase(repository, commitWithData)
    }

    router.post(
        '/upload',
        wrap(
            async (req: express.Request, res: express.Response, next: express.NextFunction): Promise<void> => {
                const { repository, commit } = req.query
                checkRepository(repository)
                checkCommit(commit)

                const filename = path.join(STORAGE_ROOT, 'uploads', uuid.v4())
                const output = fs.createWriteStream(filename)

                try {
                    const elements = readGzippedJsonElements(req)
                    const lsifElements = validateIfEnabled(elements)
                    const stringifiedLines = stringifyJsonLines(lsifElements)
                    await pipeline(Readable.from(stringifiedLines), createGzip(), output)
                } catch (e) {
                    throw Object.assign(e, { status: 422 })
                }

                // Enqueue convert job
                logger.info('enqueueing convert job', { repository, commit })
                await queue.enqueue('lsif', 'convert', [repository, commit, filename])
                res.json(null)
            }
        )
    )

    router.post(
        '/exists',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { repository, commit, file } = req.query
                checkRepository(repository)
                checkCommit(commit)

                const db = await loadDatabase(repository, commit)
                if (!db) {
                    res.json(false)
                    return
                }

                // If filename supplied, ensure we have data for it
                const result = file ? await db.exists(file) : true
                res.json(result)
            }
        )
    )

    router.post(
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

                const db = await loadDatabase(repository, commit)
                if (!db) {
                    throw Object.assign(new Error(`No LSIF data available for ${repository}@${commit}.`), {
                        status: 404,
                    })
                }

                res.json(await db[cleanMethod](path, position))
            }
        )
    )

    return router
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

// Initialize logger
const appLogger = createLogger('lsif-server')

// Run app!
main(appLogger).catch(error => {
    appLogger.error('failed to start process', { error })
    appLogger.on('finish', () => process.exit(1))
    appLogger.end()
})
