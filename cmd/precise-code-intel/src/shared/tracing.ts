import { createSilentLogger, logCall } from './logging'
import { ERROR } from 'opentracing/lib/ext/tags'
import { initTracerFromEnv } from 'jaeger-client'
import { Logger } from 'winston'
import { Span, Tracer } from 'opentracing'

/**
 * A bag of logging and tracing instances passed around a current
 * HTTP request or upload conversion.
 */
export interface TracingContext {
    /** The current tagged logger instance. Optional for testing. */
    logger?: Logger

    /** The current opentracing span. Optional for testing. */
    span?: Span
}

/**
 * Add tags to the logger and span. Returns a new context.
 *
 * @param ctx The tracing context.
 * @param tags The tags to add to the logger and span.
 */
export function addTags(
    { logger = createSilentLogger(), span = new Span() }: TracingContext,
    tags: { [name: string]: unknown }
): TracingContext {
    return { logger: logger.child(tags), span: span.addTags(tags) }
}

/**
 * Logs an event to the span of The tracing context, if its defined.
 *
 * @param ctx The tracing context.
 * @param event The name of the event.
 * @param pairs The values to log.
 */
export function logSpan(
    { span = new Span() }: TracingContext,
    event: string,
    pairs: { [name: string]: unknown }
): void {
    span.log({ event, ...pairs })
}

/**
 * Create a distributed tracer.
 *
 * @param serviceName The name of the process.
 * @param configuration The current configuration.
 */
export function createTracer(
    serviceName: string,
    {
        useJaeger,
    }: {
        /** Whether or not to enable Jaeger. */
        useJaeger: boolean
    }
): Tracer | undefined {
    if (useJaeger) {
        const config = {
            serviceName,
            sampler: {
                type: 'const',
                param: 1,
            },
        }

        return initTracerFromEnv(config, {})
    }

    return undefined
}

/**
 * Trace an operation.
 *
 * @param name The log message to output.
 * @param parent The parent span instance.
 * @param f The operation to perform.
 */
export async function traceCall<T>(name: string, parent: Span, f: (span: Span) => Promise<T> | T): Promise<T> {
    const span = parent.tracer().startSpan(name, { childOf: parent })

    try {
        return await f(span)
    } catch (error) {
        span.setTag(ERROR, true)
        span.log({
            event: 'error',
            'error.object': error,
            stack: error.stack,
            message: error.message,
        })

        throw error
    } finally {
        span.finish()
    }
}

/**
 * Log and trace the execution of a function.
 *
 * @param ctx The tracing context.
 * @param name The name of the span and text of the log message.
 * @param f The function to invoke.
 */
export function logAndTraceCall<T>(
    { logger = createSilentLogger(), span = new Span() }: TracingContext,
    name: string,
    f: (ctx: TracingContext) => Promise<T> | T
): Promise<T> {
    return logCall(name, logger, () => traceCall(name, span, childSpan => f({ logger, span: childSpan })))
}
