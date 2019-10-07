import * as fs from 'mz/fs'
import * as path from 'path'
import bodyParser from 'body-parser'
import exitHook from 'async-exit-hook'
import express from 'express'
import promBundle from 'express-prom-bundle'
import uuid from 'uuid'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { connectionCacheCapacityGauge, documentCacheCapacityGauge, resultChunkCacheCapacityGauge } from './metrics'
import { createDatabaseFilename, ensureDirectory, logErrorAndExit, readEnvInt } from './util'
import { createGzip } from 'mz/zlib'
import { createPostgresConnection } from './connection'
import { Database, tryCreateDatabase } from './database.js'
import { Edge, Vertex } from 'lsif-protocol'
import { identity } from 'lodash'
import { pipeline as _pipeline, Readable } from 'stream'
import { promisify } from 'util'
import { Queue, Scheduler } from 'node-resque'
import { readGzippedJsonElements, stringifyJsonLines, validateLsifElements } from './input'
import { wrap } from 'async-middleware'
import { XrepoDatabase } from './xrepo'
import { waitForConfiguration } from './config'

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
 * Whether or not to log a message when the HTTP server is ready and listening.
 */
const LOG_READY = process.env.DEPLOY_TYPE === 'dev'

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Whether or not to disable input validation. Validation is enabled by default.
 */
const DISABLE_VALIDATION = process.env.DISABLE_VALIDATION === 'true'

const validateIfEnabled: (data: AsyncIterable<unknown>) => AsyncIterable<Vertex | Edge> = DISABLE_VALIDATION
    ? identity
    : validateLsifElements

/**
 * Runs the HTTP server which accepts LSIF dump uploads and responds to LSIF requests.
 */
async function main(): Promise<void> {
    // Read configuration from frontend
    const fetchConfiguration = await waitForConfiguration()

    // Update cache capacities on startup
    connectionCacheCapacityGauge.set(CONNECTION_CACHE_CAPACITY)
    documentCacheCapacityGauge.set(DOCUMENT_CACHE_CAPACITY)
    resultChunkCacheCapacityGauge.set(RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await ensureDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create queue to publish jobs for worker
    const queue = await setupQueue()

    // Prepare caches
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_CAPACITY)
    const resultChunkCache = new ResultChunkCache(RESULT_CHUNK_CACHE_CAPACITY)

    // Create cross-repo database
    const connection = await createPostgresConnection(fetchConfiguration())
    const xrepoDatabase = new XrepoDatabase(connection)

    const loadDatabase = async (
        gitserverUrls: string[],
        repository: string,
        commit: string
    ): Promise<Database | undefined> => {
        // Try to construct database for the exact commit
        const database = await tryCreateDatabase(
            STORAGE_ROOT,
            xrepoDatabase,
            connectionCache,
            documentCache,
            resultChunkCache,
            repository,
            commit,
            createDatabaseFilename(STORAGE_ROOT, repository, commit)
        )
        if (database) {
            return database
        }

        // Determine the closest commit that we actually have LSIF data for. If the commit is
        // not tracked, then commit data is requested from gitserver and insert the ancestors
        // data for this commit.
        const commitWithData = await xrepoDatabase.findClosestCommitWithData(repository, commit, gitserverUrls)
        if (!commitWithData) {
            return undefined
        }

        // Try to construct a database for the approximate commit
        return tryCreateDatabase(
            STORAGE_ROOT,
            xrepoDatabase,
            connectionCache,
            documentCache,
            resultChunkCache,
            repository,
            commitWithData,
            createDatabaseFilename(STORAGE_ROOT, repository, commitWithData)
        )
    }

    const app = express()
    app.use(errorHandler)
    app.get('/ping', (_, res) => res.send('ok'))
    app.get('/healthz', (_, res) => res.send('ok'))
    app.use(promBundle({}))

    app.post(
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

                const db = await loadDatabase(fetchConfiguration().gitServers, repository, commit)
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

                const db = await loadDatabase(fetchConfiguration().gitServers, repository, commit)
                if (!db) {
                    throw Object.assign(new Error(`No LSIF data available for ${repository}@${commit}.`), {
                        status: 404,
                    })
                }

                res.json(await db[cleanMethod](path, position))
            }
        )
    )

    app.listen(HTTP_PORT, () => {
        if (LOG_READY) {
            console.log(`Listening for HTTP requests on port ${HTTP_PORT}`)
        }
    })
}

/**
 * Connect and start an active connection to the worker queue. We also run a
 * node-resque scheduler on each server instance, as these are guaranteed to
 * always be up with a responsive system. The schedulers will do their own
 * master election via a redis key and will check for dead workers attached
 * to the queue.
 */
async function setupQueue(): Promise<Queue> {
    const [host, port] = REDIS_ENDPOINT.split(':', 2)

    const connectionOptions = {
        host,
        port: parseInt(port, 10),
        namespace: 'lsif',
    }

    const queue = new Queue({ connection: connectionOptions })
    queue.on('error', logErrorAndExit)
    await queue.connect()
    exitHook(() => queue.end())

    const scheduler = new Scheduler({ connection: connectionOptions })
    scheduler.on('error', logErrorAndExit)
    await scheduler.connect()
    exitHook(() => scheduler.end())
    await scheduler.start()

    return queue
}

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 */
function errorHandler(err: any, req: express.Request, res: express.Response, next: express.NextFunction): void {
    if (err && err.status) {
        res.status(err.status).send({ message: err.message })
        return
    }

    console.error(err)
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

main().catch(logErrorAndExit)
