import * as path from 'path'
import exitHook from 'async-exit-hook'
import express from 'express'
import morgan from 'morgan'
import promBundle from 'express-prom-bundle'
import { createDirectory, logErrorAndExit, readEnvInt } from './util'
import { JobsHash, Worker, Job } from 'node-resque'
import { XrepoDatabase } from './xrepo'
import { ConnectionCache } from './cache'
import { CONNECTION_CACHE_CAPACITY_GAUGE, JOB_EVENTS_COUNTER, JOB_DURATION_HISTOGRAM } from './metrics'
import { createConvertJob, JobClasses } from './jobs'
import { JobMeta, RealWorker } from './queue'

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
    const filename = path.join(STORAGE_ROOT, 'xrepo.db')
    const xrepoDatabase = new XrepoDatabase(connectionCache, filename)

    // Start metrics server
    startMetricsServer()

    // Create worker and start processing jobs
    await startWorker({
        'convert': createConvertJob(STORAGE_ROOT, xrepoDatabase),
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

    // Create worker and attach loggers to the interesting events
    const worker = new Worker({ connection: connectionOptions, queues: ['lsif'] }, jobs) as RealWorker
    worker.on('start', () => console.log('Worker started'))
    worker.on('end', () => console.log('Worker ended'))
    worker.on('poll', () => console.log('Polling queue'))
    worker.on('ping', () => console.log('Pinging queue'))
    worker.on('cleaning_worker', (worker: string, pid: string) => console.log(`Cleaning old worker ${worker}:${pid}`))
    worker.on('error', logErrorAndExit)

    let end: (() => void) | undefined

    worker.on('job', (_: string, job: Job<any> & JobMeta) => {
        console.log(`Working on job ${JSON.stringify(job)}`)
        end = JOB_DURATION_HISTOGRAM.labels(job.class).startTimer()
    })

    worker.on('success', (_: string, job: Job<any> & JobMeta, result: any) => {
        console.log(`Completed job ${JSON.stringify(job)} >> ${result}`)
        JOB_EVENTS_COUNTER.labels(job.class, 'success').inc()
        end && end()
        end = undefined
    })

    worker.on('failure', (_: string, job: Job<any> & JobMeta, failure: any) => {
        console.log(`Failed job ${JSON.stringify(job)} >> ${failure}`)
        JOB_EVENTS_COUNTER.labels(job.class, 'failure').inc()
        end && end()
        end = undefined
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
    const app = express()
    app.use(morgan('tiny'))
    app.get('/healthz', (_, res) => res.send('ok'))
    app.use(promBundle({}))

    app.listen(WORKER_METRICS_PORT, () => {
        if (LOG_READY) {
            console.log(`Listening for HTTP requests on port ${WORKER_METRICS_PORT}`)
        }
    })
}

main().catch(logErrorAndExit)
