import * as fs from 'mz/fs'
import * as path from 'path'
import * as zlib from 'mz/zlib'
import exitHook from 'async-exit-hook'
import express from 'express'
import morgan from 'morgan'
import promBundle from 'express-prom-bundle'
import uuid from 'uuid'
import { convertLsif } from './conversion'
import { createDatabaseFilename, createDirectory, logErrorAndExit, readEnvInt } from './util'
import { JobsHash, Worker } from 'node-resque'
import { XrepoDatabase } from './xrepo'
import { ConnectionCache } from './cache'
import { CONNECTION_CACHE_CAPACITY_GAUGE } from './metrics'

/**
 * Which port to run the worker metrics server on. Defaults to 3187.
 */
const WORKER_METRICS_PORT = readEnvInt('WORKER_METRICS_PORT', 3187)

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
const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 1024 * 1024 * 1024)

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
    CONNECTION_CACHE_CAPACITY_GAUGE.set(CONNECTION_CACHE_CAPACITY)

    // Ensure storage roots exist
    await createDirectory(STORAGE_ROOT)
    await createDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await createDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create backend
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const filename = path.join(STORAGE_ROOT, 'correlation.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)

    const jobFunctions = {
        convert: createConvertJob(xrepoDatabase),
    }

    // Start metrics server
    startMetricsServer()

    // Create worker and start processing jobs
    startWorker(jobFunctions)

    if (LOG_READY) {
        console.log('Listening for uploads')
    }
}

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the correlation database for this dump.
 *
 * @param xrepoDatabase The correlation database.
 */
function createConvertJob(
    xrepoDatabase: XrepoDatabase
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        console.log(`Converting ${repository}@${commit}`)

        const input = fs.createReadStream(filename).pipe(zlib.createGunzip())
        const tempFile = path.join(STORAGE_ROOT, 'tmp', uuid.v4())

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile)

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, createDatabaseFilename(STORAGE_ROOT, repository, commit))

            // Add the new database to the correlation db
            await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
        } catch (e) {
            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }

        // Remove input
        await fs.unlink(filename)
    }
}

/**
 * Connect to redis and begin processing work with the given hash of job functions.
 *
 * @param jobFunctions An object whose values are the functions to execute for a job name matching its key.
 */
async function startWorker(jobFunctions: { [name: string]: (...args: any[]) => Promise<any> }): Promise<void> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
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
    worker.start().catch(logErrorAndExit)
}

/**
 * Create an express server that only has /ping and /metric endpoints.
 */
function startMetricsServer(): void {
    const app = express()
    app.use(morgan('tiny'))
    app.get('/ping', (_, res) => res.send({ pong: 'pong' }))
    app.use(promBundle({}))

    app.listen(WORKER_METRICS_PORT, () => {
        if (LOG_READY) {
            console.log(`Listening for HTTP requests on port ${WORKER_METRICS_PORT}`)
        }
    })
}

main().catch(logErrorAndExit)
