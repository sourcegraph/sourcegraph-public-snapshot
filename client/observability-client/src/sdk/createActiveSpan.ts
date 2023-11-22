import {
    context,
    trace,
    type Tracer,
    ROOT_CONTEXT,
    type SpanOptions,
    type Context,
    type Span,
    type TimeInput,
} from '@opentelemetry/api'

export interface ActiveSpanConfig extends SpanOptions {
    name: string
    startTime?: TimeInput
    endTime?: TimeInput
    parentSpan?: Span
    context?: Context
}

/**
 * Creates span, links to a parent span, calls callback in the new span context.
 * A helper to use with the Web Performance API where the `endTime` is often available right away.
 *
 * See https://opentelemetry.io/docs/instrumentation/js/instrumentation/#create-nested-spans
 */
export function createActiveSpan<F extends (span: Span) => ReturnType<F>>(
    tracer: Tracer,
    config: ActiveSpanConfig,
    callback: F
): ReturnType<F> {
    const { name, startTime, parentSpan, context: spanContext = ROOT_CONTEXT, ...restSpanOptions } = config

    const resultContext = parentSpan ? trace.setSpan(context.active(), parentSpan) : spanContext

    return tracer.startActiveSpan<F>(
        name,
        {
            startTime,
            ...restSpanOptions,
        },
        resultContext,
        callback
    )
}
