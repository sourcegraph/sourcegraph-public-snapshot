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
import { createLogger } from './logging'
import { createPostgresConnection } from './connection'
import { Database } from './database.js'
import { Edge, Vertex } from 'lsif-protocol'
import { updateCommits } from './commits'
import { identity } from 'lodash'
import { logger as loggingMiddleware } from 'express-winston'
import { Logger } from 'winston'
import { pipeline as _pipeline, Readable } from 'stream'
import { promisify } from 'util'
import { Queue, Scheduler } from 'node-resque'
import { readGzippedJsonElements, stringifyJsonLines, validateLsifElements } from './input'
import { wrap } from 'async-middleware'
import { XrepoDatabase } from './xrepo'
import { monitor, MonitoringContext } from './monitoring'
import { Tracer, Span } from 'opentracing'
import { default as tracingMiddleware } from 'express-opentracing'
import { waitForConfiguration, ConfigurationFetcher } from './config'
import { createTracer } from './tracing'

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
    // Read configuration from frontend
    const configurationFetcher = await waitForConfiguration(logger)

    // Configure distributed tracing
    const tracer = createTracer('lsif-server', configurationFetcher())

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
    app.use(promBundle({}))

    // Register endpoints
    app.use(metaEndpoints())
    app.use(await lsifEndpoints(queue, configurationFetcher, logger, tracer))

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
 * @param logger The logger instance.
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
 * @param configurationFetcher A function that returns the current configuration.
 * @param logger The logger instance.
 * @param tracer The tracer instance.
 */
async function lsifEndpoints(
    queue: Queue,
    configurationFetcher: ConfigurationFetcher,
    logger: Logger,
    tracer: Tracer | undefined
): Promise<express.Router> {
    const router = express.Router()

    // Create cross-repo database
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_CAPACITY)
    const resultChunkCache = new ResultChunkCache(RESULT_CHUNK_CACHE_CAPACITY)

    // Create cross-repo database
    const connection = await createPostgresConnection(configurationFetcher(), logger)
    const xrepoDatabase = new XrepoDatabase(connection)

    /**
     * Return the name of the database file for the given repository and commit
     * if it exists, returning undefined otherwise.
     *
     * @param repository The repository name.
     * @param commit The commit.
     */
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

    /**
     * Create a database for the given repository and commit if the underlying
     * file exists.
     *
     * @param repository The repository name.
     * @param commit The commit.
     */
    const tryCreateDatabase = async (repository: string, commit: string): Promise<Database | undefined> => {
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

    /**
     * Create a database for the given repository and a commit nearest to the
     * given commit or which we have LSIF data commit. This may request more
     * data from gitserver on-demand.
     *
     * @param gitserverUrls The set of ordered gitserver urls.
     * @param repository The repository name.
     * @param commit The commit.
     * @param ctx The monitoring context.
     */
    const createDatabase = async (
        gitserverUrls: string[],
        repository: string,
        commit: string,
        ctx: MonitoringContext
    ): Promise<Database | undefined> => {
        // Try to construct database for the exact commit
        const database = await tryCreateDatabase(repository, commit)
        if (database) {
            return database
        }

        // Request updated commit data from gitserver if this commit isn't
        // already tracked. This will pull back ancestors for this commit
        // up to a certain (configurable) depth and insert them into the
        // cross-repository database. This populates the necessary data for
        // the following query.
        await monitor(ctx, 'updating commits for repo', (ctx: MonitoringContext) =>
            updateCommits(gitserverUrls, xrepoDatabase, repository, commit, ctx)
        )

        // Determine the closest commit that we actually have LSIF data for
        const commitWithData = await monitor(ctx, 'querying closest commit with LISF data', () =>
            xrepoDatabase.findClosestCommitWithData(repository, commit)
        )

        if (!commitWithData) {
            return undefined
        }

        // Try to construct a database for the approximate commit
        return tryCreateDatabase(repository, commitWithData)
    }

    /**
     * Create a monitoring context from the request logger and tracing span
     * tagged with the currently request repository and commit.
     *
     * @param req The express request.
     * @param repository The repository name.
     * @param commit The commit.
     */
    const createMonitoringContext = (
        req: express.Request & { span?: Span },
        repository: string,
        commit: string
    ): MonitoringContext => ({
        // Tag logger with target repo/commit
        logger: logger.child({ repository, commit }),
        // Pull span from request (injected by middleware)
        span: req.span && req.span,
    })

    router.post(
        '/upload',
        wrap(
            async (
                req: express.Request & { span?: Span },
                res: express.Response,
                next: express.NextFunction
            ): Promise<void> => {
                const { repository, commit } = req.query
                checkRepository(repository)
                checkCommit(commit)

                const ctx = createMonitoringContext(req, repository, commit)
                const filename = path.join(STORAGE_ROOT, 'uploads', uuid.v4())
                const output = fs.createWriteStream(filename)

                try {
                    await monitor(ctx, 'uploading dump', async () => {
                        const elements = readGzippedJsonElements(req)
                        const lsifElements = validateIfEnabled(elements)
                        const stringifiedLines = stringifyJsonLines(lsifElements)
                        await pipeline(Readable.from(stringifiedLines), createGzip(), output)
                    })
                } catch (e) {
                    throw Object.assign(e, { status: 422 })
                }

                // Enqueue convert job
                logger.debug('enqueueing convert job', { repository, commit })
                await queue.enqueue('lsif', 'convert', [repository, commit, filename])
                res.json(null)
            }
        )
    )

    router.post(
        '/exists',
        wrap(
            async (req: express.Request & { span?: Span }, res: express.Response): Promise<void> => {
                const { repository, commit, file } = req.query
                checkRepository(repository)
                checkCommit(commit)

                const ctx = createMonitoringContext(req, repository, commit)
                const db = await monitor(ctx, 'creating database', ctx => createDatabase(configurationFetcher().gitServers, repository, commit, ctx))
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
            async (req: express.Request & { span?: Span }, res: express.Response): Promise<void> => {
                const { repository, commit } = req.query
                const { path, position, method } = req.body
                checkRepository(repository)
                checkCommit(commit)
                checkMethod(method, ['definitions', 'references', 'hover'])
                const cleanMethod = method as 'definitions' | 'references' | 'hover'

                const ctx = createMonitoringContext(req, repository, commit)
                const db = await monitor(ctx, 'creating database', ctx => createDatabase(configurationFetcher().gitServers, repository, commit, ctx))
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
        throw Object.assign(new Error(`Must specify the commit as a 40 character hash ${commit}`), { status: 400 })
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
