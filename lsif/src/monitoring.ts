import { Logger } from 'winston'
import { Span } from 'opentracing'
import { trace } from './tracing'
import { log } from './logging'

/**
 * A bag of logging and tracing instances passed around a current
 * HTTP request or worker job.
 */
export interface MonitoringContext {
    /**
     * The current tagged logger instance. Optional for testing.
     */
    logger?: Logger

    /**
     * The current opentracing span. Optional for testing.
     */
    span?: Span
}

/**
 * Logs an event to the span of the monitoring context, if its defined.
 *
 * @param ctx The monitoring context.
 * @param event The name of the event.
 * @param pairs The values to log.
 */
export function logSpan(ctx: MonitoringContext, event: string, pairs: { [K: string]: any }): void {
    if (ctx.span) {
        ctx.span.log({ event, ...pairs })
    }
}

/**
 * Log and trace the execution of a function.
 *
 * @param ctx The monitoring context.
 * @param name The name of the span and text of the log message.
 * @param f The function to invoke.
 */
export function monitor<T>(
    ctx: MonitoringContext,
    name: string,
    f: (ctx: MonitoringContext) => Promise<T> | T
): Promise<T> | T {
    const wrapTrace = (): Promise<T> | T => {
        if (ctx.span) {
            return trace(name, ctx.span, (span: Span) => f({ logger: ctx.logger, span }))
        }

        return f(ctx)
    }

    return ctx.logger ? log(name, ctx.logger, wrapTrace) : wrapTrace()
}
