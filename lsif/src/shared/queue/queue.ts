import Bull, { Queue, Job, JobOptions } from 'bull'
import { Span, Tracer, FORMAT_TEXT_MAP } from 'opentracing'
import { Logger } from 'winston'

/**
 * The name of the task queue for this process.
 */
export const QUEUE_NAME = 'lsif'

/**
 * The key prefix used by Bull. If the queue setup code changes, this may need
 * to change as well.
 */
export const QUEUE_PREFIX = `bull:${QUEUE_NAME}:`

/**
 * The names of queues as defined in Bull.
 */
export type QueueTypes = 'active' | 'failed' | 'completed' | 'wait' | 'delayed'

/**
 * The names of job states in the API.
 */
export type ApiJobState = 'processing' | 'errored' | 'completed' | 'queued' | 'scheduled'

/**
 * A mapping from job states to queue names.
 */
export const queueTypes = new Map<ApiJobState, QueueTypes>([
    ['processing', 'active'],
    ['errored', 'failed'],
    ['completed', 'completed'],
    ['queued', 'wait'],
    ['scheduled', 'delayed'],
])

/**
 * A mapping from queue names to job states.
 */
export const statesByQueue = new Map(
    Array.from(queueTypes.entries()).map(([state, queue]) => [queue, state] as [QueueTypes, ApiJobState])
)

/**
 * Creates a queue instance.
 *
 * @param endpoint The host:port redis address.
 * @param logger The logger instance.
 */
export function createQueue(endpoint: string, logger: Logger): Queue {
    const [host, port] = endpoint.split(':', 2)

    const redis = {
        host,
        port: parseInt(port, 10),
    }

    const queue = new Bull(QUEUE_NAME, { redis })
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
            if (job.every === intervalMs) {
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
    await enqueue(queue, name, args, { repeat: { every: intervalMs } })
}
