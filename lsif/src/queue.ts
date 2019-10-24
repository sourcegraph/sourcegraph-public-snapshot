import Bull, { Queue, Job, JobOptions } from 'bull'
import { Span, Tracer, FORMAT_TEXT_MAP } from 'opentracing'
import { Logger } from 'winston'

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
