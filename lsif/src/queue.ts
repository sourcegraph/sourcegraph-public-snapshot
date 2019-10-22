import Bull, { Queue, Job, JobOptions } from 'bull'
import { Span, Tracer, FORMAT_TEXT_MAP } from 'opentracing'
import { Logger } from 'winston'
import { promisify } from 'util'
import { chunk } from 'lodash'

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
 * The representation of a job as returned by the API.
 */
export interface JsonJob {
    id: string
    name: string
    args: object
    status: string
    progress: number
    timestamp: string
    finishedOn: string | null
    processedOn: string | null
    failedReason: string | null
    stacktrace: string[] | null
}

export type JobStatus = 'active' | 'waiting' | 'delayed' | 'completed' | 'failed'

export const queueTypes = new Map<string, JobStatus>([
    ['active', 'active'],
    ['queued', 'waiting'],
    ['scheduled', 'delayed'],
    ['completed', 'completed'],
    ['failed', 'failed'],
])

/**
 * Format a job to return from the API.
 *
 * @param job The job to format.
 * @param status The job's status.
 */
export const formatJob = (job: Job, status: string): JsonJob => {
    const payload = job.toJSON()

    return {
        id: `${payload.id}`,
        name: payload.name,
        args: payload.data.args,
        status,
        progress: payload.progress,
        timestamp: new Date(payload.timestamp).toISOString(),
        finishedOn: payload.finishedOn ? new Date(payload.finishedOn).toISOString() : '',
        processedOn: payload.processedOn ? new Date(payload.processedOn).toISOString() : '',
        failedReason: payload.failedReason,
        stacktrace: payload.stacktrace,
    }
}

/**
 * Format a job to return from the API.
 *
 * @param values A map of values composing the job.
 * @param status The job's status.
 */
const formatJobFromMap = (values: Map<string, string>, status: string): JsonJob => {
    const rawStacktrace = values.get('stacktrace')
    const rawFinishedOn = values.get('finishedOn')
    const rawProcessedOn = values.get('processedOn')

    return {
        id: values.get('id') || '',
        name: values.get('name') || '',
        args: JSON.parse(values.get('data') || '{}').args || {},
        status,
        progress: parseInt(values.get('progress') || '', 10),
        timestamp: new Date(parseInt(values.get('timestamp') || '')).toISOString(),
        finishedOn: (rawFinishedOn && new Date(parseInt(rawFinishedOn, 10)).toISOString()) || null,
        processedOn: (rawProcessedOn && new Date(parseInt(rawProcessedOn, 10)).toISOString()) || null,
        failedReason: values.get('failedReason') || null,
        stacktrace: (rawStacktrace && (JSON.parse(rawStacktrace) as any[])) || null,
    }
}

/**
 * Convert an hgetall response `[k1, v1, k2, v2, ...]` into a map `{k1 -> v1, k2 -> v2, ...}`.
 *
 * @param payload Response from hgetall.
 */
const hgetAllToMap = (payload: string[]): Map<string, string> =>
    new Map<string, string>(chunk(payload, 2) as [string, string][])

/**
 * A Lua script evaluated in Redis to return the jobs in the given queue
 * matching the given query string. This script has the following input:
 *
 *  - KEYS[1]: redis key prefix (e.g. bull:lsif:)
 *  - KEYS[2]: the name of the queues
 *
 *  - ARGV[1]: the search query
 *  - ARGV[2]: start index to scan (inclusive)
 *  - ARGV[3]: end index to scan (inclusive)
 */
const jobSearchScript = `
    local function textMatches(needle, haystack)
        for term in string.gmatch(needle, '%S+') do
            if string.find(haystack, term, 1, true) == nil then
                return false
            end
        end

        return true
    end

    local function jobMatches(key)
        for _, field in pairs({'data'}) do
            if textMatches(ARGV[1], redis.call('HGET', key, field)) then
                return true
            end
        end

        return false
    end

    local command = 'ZRANGE'
    if KEYS[2] == 'active' then
        command = 'LRANGE'
    end

    local matching = {}
    for _, v in pairs(redis.call(command, KEYS[1] .. KEYS[2], ARGV[2], ARGV[3])) do
        if jobMatches(KEYS[1] .. v) then
            table.insert(matching, redis.call('HGETALL', KEYS[1] .. v))
        end
    end

    return matching
`

/**
 * Return a list of JSON-encoded jobs with the given status and that contain the
 * given search term.
 *
 * @param queue The job queue.
 * @param queueName The queue name.
 * @param search The search query.
 * @param start The start index (inclusive).
 * @param end The end index (inclusive).
 */
export async function searchJobs(
    queue: Queue,
    queueName: string,
    search: string,
    start: number,
    end: number
): Promise<JsonJob[]> {
    const evalCommand = promisify(queue.client.eval.bind(queue.client)) as (
        lua: string,
        numberOfKeys: number,
        keysAndArgs: any[]
    ) => Promise<string[][]>

    const keys = ['bull:lsif:', queueName]
    const args = [search, start, end]

    const jobs = []
    for (const payload of await evalCommand(jobSearchScript, keys.length, [...keys, ...args])) {
        jobs.push(formatJobFromMap(hgetAllToMap(payload), status))
    }

    return jobs
}
