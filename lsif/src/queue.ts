import Bull, { Queue, Job, JobOptions } from 'bull'
import { Span, Tracer, FORMAT_TEXT_MAP } from 'opentracing'
import { Logger } from 'winston'
import { promisify } from 'util'
import { chunk } from 'lodash'
import { ApiJob, formatJobFromMap, formatJob } from './api-job'
import * as fs from 'mz/fs'
import { readEnvInt } from './util'
import { Redis } from 'ioredis'

/**
 * The maximum number of jobs that will be searched on a search operation.
 */
const MAX_JOB_SEARCH = readEnvInt('MAX_JOB_SEARCH', 10000)

/**
 * The key prefix used by Bull. If the queue setup code changes, this may need
 * to change as well.
 */
const QUEUE_PREFIX = 'bull:lsif:'

/**
 * The names of queues as defined in Bull.
 */
export type QueueTypes = 'active' | 'wait' | 'delayed' | 'completed' | 'failed'

/**
 * A mapping from job statuses to queue names.
 */
export const queueTypes = new Map<string, QueueTypes>([
    ['active', 'active'],
    ['queued', 'wait'],
    ['scheduled', 'delayed'],
    ['completed', 'completed'],
    ['failed', 'failed'],
])

/**
 * Creates a queue instance.
 *
 * @param name The name of the queue.
 * @param endpoint The host:port redis address.
 * @param logger The logger instance.
 */
export function createQueue(name: string, endpoint: string, logger: Logger): Queue {
    const [host, port] = endpoint.split(':', 2)

    const redis = {
        host,
        port: parseInt(port, 10),
    }

    const queue = new Bull(name, { redis })
    queue.on('error', (error: Error) => logger.error('queue error', { error }))
    queue.on('global:stalled', (id: string) => logger.error('job stalled', { jobId: id }))

    return queue
}

/**
 * Enqueue a job to be run by a worker.
 *
 * @param queue The job queue.
 * @param name The name of the job class.
 * @param args The job arguments.
 * @param opts The job options.
 * @param tracer The tracer instance.
 * @param span The parent span.
 */
export const enqueue = (
    queue: Queue,
    name: string,
    args: object,
    opts: JobOptions,
    tracer?: Tracer,
    span?: Span
): Promise<Job> => {
    const tracing = {}
    if (tracer && span) {
        tracer.inject(span, FORMAT_TEXT_MAP, tracing)
    }

    return queue.add(name, { args, tracing }, opts)
}

/**
 * Schedule a job to be invoked on an interval. If this job was previously
 * scheduled with a different interval, the old instance is first unscheduled.
 *
 * @param queue The job queue.
 * @param name The name of the job class.
 * @param args The job arguments.
 * @param intervalMs How frequently to run the job.
 */
export const ensureOnlyRepeatableJob = async (
    queue: Queue,
    name: string,
    args: object,
    intervalMs: number
): Promise<void> => {
    const keys = []
    for (const job of await queue.getRepeatableJobs()) {
        if (job.name === name) {
            // Job already scheduled with correct interval
            if (job.every === intervalMs * 1000) {
                return
            }

            keys.push(job.key)
        }
    }

    for (const key of keys) {
        // Remove old scheduled jobs with different intervals
        await queue.removeRepeatableByKey(key)
    }

    // Schedule job with correct interval
    await enqueue(queue, name, args, { repeat: { every: intervalMs * 1000 } })
}

/**
 * Return a list of JSON-encoded jobs with the given status.
 *
 * @param queue The job queue.
 * @param status The target job status.
 * @param start The start index (inclusive).
 * @param end The end index (inclusive).
 */
export async function sliceJobs(
    queue: Queue,
    status: string,
    start: number,
    end: number
): Promise<{ jobs: ApiJob[]; totalCount: number }> {
    const queueName = queueTypes.get(status)
    if (!queueName) {
        throw new Error(`Unknown job status ${status}`)
    }

    const rawJobs = await queue.getJobs([queueName], start, end)
    const jobs = rawJobs.map(job => formatJob(job, status))
    const totalCount = (await queue.getJobCountByTypes([queueName])) as never
    return { jobs, totalCount }
}

/**
 * Translate a redis hgetall response (`[k1, v2, k2, v1, ...]`) into a map (`{k1 -> v1, k2 -> v2, ...}`).
 *
 * @param payload The hgetall response payload.
 */
const hgetallResponseToMap = (payload: string[]): Map<string, string> =>
    new Map<string, string>(chunk(payload, 2) as [string, string][])

/**
 * Return a list of JSON-encoded jobs with the given status and that contain the
 * given search term.
 *
 * @param queue The job queue.
 * @param status The target job status.
 * @param search The search query.
 * @param limit The maximum number of results to return.
 */
export async function searchJobs(queue: Queue, status: string, search: string, limit: number): Promise<ApiJob[]> {
    const queueName = queueTypes.get(status)
    if (!queueName) {
        throw new Error(`Unknown job status ${status}`)
    }

    const args = [QUEUE_PREFIX, queueName, search, limit, MAX_JOB_SEARCH]
    const payloads = await evalSearch(queue.client, args)
    return payloads.map(payload => formatJobFromMap(hgetallResponseToMap(payload), status))
}

/**
 * The type of the promisified Redis `script` command.
 */
type scriptCommandType = (subcommand: string, lua: string) => Promise<string>

/**
 * The type of the promisified Redis `evalsha` command (specific to the search-jobs script).
 */
type evalshaCommandType = (sha: string, n: number, args: any[]) => Promise<string[][]>

/**
 * The ioredis library has some weird types that only take callbacks. This converts a
 * command on the client to take promises (but doesn't do so in such a nice typesafe way).
 * All ioredis commands take a slice of args, which isn't so typesafe to begin wth.
 *
 * @param client The redis client.
 * @param name The command to promisify.
 */
const promisfyRedis = <T>(client: Redis, name: keyof Redis): T =>
    promisify((client[name] as (...args: any[]) => void).bind(client)) as never

/**
 * The SHA1 hash of the search-jobs Lua script in Redis. This is populated by first
 * loading the script but not executing it. Evaluations of the script can send only
 * the hash instead of the entire text, which saves on bandwidth and parsing.
 */
let sha: string | undefined

/**
 * Load the search-jobs.lua script into Redis, if necessary, and evaluate it with the
 * given arguments.
 *
 * @param client The redis client.
 * @param args The arguments to supply to the script.
 */
async function evalSearch(client: Redis, args: any[]): Promise<string[][]> {
    if (!sha) {
        const script = await fs.readFile(`${__dirname}/search-jobs.lua`)
        const newSha = await promisfyRedis<scriptCommandType>(client, 'script')('load', script.toString())

        // eslint-disable-next-line require-atomic-updates
        sha = newSha
    }

    return promisfyRedis<evalshaCommandType>(client, 'evalsha')(sha, 2, args)
}
