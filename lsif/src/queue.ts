import Bull, { Queue, Job } from 'bull'
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
        namespace: `lsif_${name}`,
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
 * @param args The job arguments.
 * @param tracer The tracer instance.
 * @param span The parent span.
 */
export const enqueue = (queue: Queue, args: object, tracer?: Tracer, span?: Span): Promise<Job> => {
    const tracing = {}
    if (tracer && span) {
        tracer.inject(span, FORMAT_TEXT_MAP, tracing)
    }

    return queue.add({ args, tracing })
}
