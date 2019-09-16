import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import promClient from 'prom-client'
import { logger, initLogger } from './logger'
import { CONNECTION_CACHE_CAPACITY_GAUGE, JOB_DURATION_HISTOGRAM, JOB_EVENTS_COUNTER } from './metrics'
import { ConnectionCache } from './cache'
import { createConvertJob, JobClasses } from './jobs'
import { createDirectory, logErrorAndExit, readEnvInt } from './util'
import { Job, JobsHash, Worker } from 'node-resque'
import { JobMeta, RealWorker } from './queue'
import { XrepoDatabase } from './xrepo'

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
const CONNECTION_CACHE_CAPACITY = readEnvInt('CONNECTION_CACHE_CAPACITY', 100)

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Runs the worker which accepts LSIF conversion jobs from node-resque.
 */
async function main(): Promise<void> {
    // Initialize logger
    initLogger('lsif-worker')

    // Update cache capacities on startup
    CONNECTION_CACHE_CAPACITY_GAUGE.set(CONNECTION_CACHE_CAPACITY)

    // Ensure storage roots exist
    await createDirectory(STORAGE_ROOT)
    await createDirectory(path.join(STORAGE_ROOT, 'tmp'))
    await createDirectory(path.join(STORAGE_ROOT, 'uploads'))

    // Create backend
    const connectionCache = new ConnectionCache(CONNECTION_CACHE_CAPACITY)
    const filename = path.join(STORAGE_ROOT, 'xrepo.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)

    // Start metrics server
    startMetricsServer()

    // Create worker and start processing jobs
    await startWorker({
        convert: createConvertJob(STORAGE_ROOT, xrepoDatabase),
    })
}

/**
 * Connect to redis and begin processing work with the given hash of job functions.
 *
 * @param jobFunctions An object whose values are the functions to execute for a job name matching its key.
 */
async function startWorker(jobFunctions: { [K in JobClasses]: (...args: any[]) => Promise<any> }): Promise<void> {
    const connectionOptions = {
        host: REDIS_HOST,
        port: REDIS_PORT,
        namespace: 'lsif',
    }

    const jobs: JobsHash = {}
    for (const [key, fn] of Object.entries(jobFunctions)) {
        jobs[key] = { perform: fn }
    }

    // Create worker and log the interesting events
    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, jobs) as RealWorker
    worker.on('start', () => logger.debug('Started worker'))
    worker.on('end', () => logger.debug('Ended worker'))
    worker.on('poll', () => logger.debug('Polling queue'))
    worker.on('ping', () => logger.debug('Pinging queue'))
    worker.on('error', logErrorAndExit)

    worker.on('cleaning_worker', (worker: string, pid: string) =>
        logger.debug('Cleaning old worker', { worker: `${worker}:${pid}` })
    )

    // Start a timer when accepting a job and end it when either
    // succeeding or failing. This is fine as we're not using a
    // multiWorker and only one job will be processed at a time.
    let end: (() => void) | undefined

    worker.on('job', (_: string, job: Job<any> & JobMeta) => {
        logger.debug('Working on job', { job })
        end = JOB_DURATION_HISTOGRAM.labels(job.class).startTimer()
    })

    worker.on('success', (_: string, job: Job<any> & JobMeta, result: any) => {
        logger.debug('Completed job', { job, result })
        JOB_EVENTS_COUNTER.labels(job.class, 'success').inc()
        if (end) {
            end()
        }
    })

    worker.on('failure', (_: string, job: Job<any> & JobMeta, failure: any) => {
        logger.debug('Failed job', { job, failure })
        JOB_EVENTS_COUNTER.labels(job.class, 'failure').inc()
        if (end) {
            end()
        }
    })

    // Start worker
    await worker.connect()
    exitHook(() => worker.end())
    worker.start().catch(logErrorAndExit)
}

/**
 * Create an express server that only has /ping and /metric endpoints.
 */
function startMetricsServer(): void {
    // Create app
    const app = express()

    // Register endpoints
    app.get('/healthz', (_, res) => res.send('ok'))
    app.get('/metrics', (_, res) => {
        res.writeHead(200, { 'Content-Type': 'text/plain' })
        res.end(promClient.register.metrics())
    })

    app.listen(WORKER_METRICS_PORT, () => {
        logger.debug('Serving worker metrics', { port: WORKER_METRICS_PORT })
    })
}

main().catch(logErrorAndExit)
