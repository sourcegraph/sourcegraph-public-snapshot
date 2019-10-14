import { Queue } from 'node-resque'
import { Span, Tracer, FORMAT_TEXT_MAP } from 'opentracing'

/**
 * Enqueue a job to be run by the worker.
 *
 * @param queue The job queue.
 * @param job The job name.
 * @param args The job arguments.
 * @param tracer The tracer instance.
 * @param span The parent span.
 */
export const enqueue = (
    queue: Queue,
    job: string,
    args: { [K: string]: any },
    tracer?: Tracer,
    span?: Span
): Promise<void> => {
    if (tracer && span) {
        const tracing = {}
        tracer.inject(span, FORMAT_TEXT_MAP, tracing)
        args.tracing = tracing
    }

    return queue.enqueue('lsif', job, [args])
}
