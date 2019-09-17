import * as fs from 'mz/fs'
import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import promBundle from 'express-prom-bundle'
import { Backend, createBackend } from './backend'
import { ConnectionCache, DocumentCache, ResultChunkCache } from './cache'
import { connectionCacheCapacityGauge, documentCacheCapacityGauge, resultChunkCacheCapacityGauge } from './metrics'
import { ensureDirectory, logErrorAndExit, readEnvInt } from './util'
import { JobsHash, Worker } from 'node-resque'

/**
 * Which port to run the worker metrics server on. Defaults to 3187.
 */
const WORKER_METRICS_PORT = readEnvInt('WORKER_METRICS_PORT', 3187)

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
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 */
async function main(): Promise<void> {
    // Update cache capacities on startup
    connectionCacheCapacityGauge.set(CONNECTION_CACHE_CAPACITY)
    documentCacheCapacityGauge.set(DOCUMENT_CACHE_CAPACITY)
    resultChunkCacheCapacityGauge.set(RESULT_CHUNK_CACHE_CAPACITY)

    // Ensure storage roots exist
    await ensureDirectory(STORAGE_ROOT)
    await ensureDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await ensureDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create backend
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const documentCache = new DocumentCache(DOCUMENT_CACHE_CAPACITY)
    const resultChunkCache = new ResultChunkCache(RESULT_CHUNK_CACHE_CAPACITY)
    const backend = await createBackend(STORAGE_ROOT, connectionCache, documentCache, resultChunkCache)

    const jobFunctions = {
        convert: createConvertJob(backend),
    }

    // Start metrics server
    startMetricsServer()

    // Create worker and start processing jobs
    await startWorker(jobFunctions)

    if (LOG_READY) {
        console.log('Listening for uploads')
    }
}

/**
 * Create a job that takes a repository, commit, and input filename as input and converts
 * the convents of the file into a SQLite database.
 *
 * @param backend The backend instance.
 */
function createConvertJob(backend: Backend): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        console.log(`Converting ${repository}@${commit}`)
        const input = fs.createReadStream(filename)
        await backend.insertDump(input, repository, commit)
        await fs.unlink(filename)
    }
}

/**
 * Connect to redis and begin processing work with the given hash of job functions.
 *
 * @param jobFunctions An object whose values are the functions to execute for a job name matching its key.
 */
async function startWorker(jobFunctions: { [name: string]: (...args: any[]) => Promise<any> }): Promise<void> {
    const [host, port] = REDIS_ENDPOINT.split(':', 2)

    const connectionOptions = {
        host,
        port: parseInt(port, 10),
        namespace: 'lsif',
    }

    const jobs: JobsHash = {}
    for (const key of Object.keys(jobFunctions)) {
        jobs[key] = { perform: jobFunctions[key] }
    }

    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, jobs)
    worker.on('error', logErrorAndExit)
    await worker.connect()
    exitHook(() => worker.end())
    await worker.start()
}

/**
 * Create an express server that only has /healthz and /metric endpoints.
 */
function startMetricsServer(): void {
    const app = express()
    app.get('/healthz', (_, res) => res.send('ok'))
    app.use(promBundle({}))

    app.listen(WORKER_METRICS_PORT, () => {
        if (LOG_READY) {
            console.log(`Listening for HTTP requests on port ${WORKER_METRICS_PORT}`)
        }
    })
}

main().catch(logErrorAndExit)
