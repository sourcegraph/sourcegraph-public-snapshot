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
import { addSpanNetworkEvents, type PerformanceEntries } from '@opentelemetry/sdk-trace-web'

export interface FinishedSpanConfig extends SpanOptions {
    name: string
    startTime?: TimeInput
    endTime?: TimeInput
    parentSpan?: Span
    context?: Context
    networkEvents?: PerformanceEntries
}

/**
 * Creates span, links to a parent span, adds network events, ends the span.
 * A helper to use with the Web Performance API where the `endTime` is often available right away.
 *
 * See https://developer.mozilla.org/en-US/docs/Web/API/Performance
 */
export function createFinishedSpan(tracer: Tracer, config: FinishedSpanConfig): Span {
    const {
        name,
        startTime,
        endTime,
        parentSpan,
        context: spanContext = ROOT_CONTEXT,
        networkEvents,
        ...restSpanOptions
    } = config

    const resultContext = parentSpan ? trace.setSpan(context.active(), parentSpan) : spanContext

    const span = tracer.startSpan(
        name,
        {
            startTime,
            ...restSpanOptions,
        },
        resultContext
    )

    if (networkEvents) {
        addSpanNetworkEvents(span, networkEvents)
    }

    span.end(endTime)

    return span
}
