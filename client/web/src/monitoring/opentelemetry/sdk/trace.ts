import { context, trace, Span, TimeInput } from '@opentelemetry/api'
import { TRACE_PARENT_HEADER } from '@opentelemetry/core'

/**
 * Parses `traceParent` - a meta property that comes from server.
 * It should be dynamically generated server side to have the server's request trace Id,
 * a parent span Id that was set on the server's request span.
 *
 * See https://opentelemetry.io/docs/instrumentation/js/getting-started/browser/
 */
export function getServerSideTraceParent(): string {
    const metaElement = Array.from(document.querySelectorAll('meta')).find(
        element => element.getAttribute('name') === TRACE_PARENT_HEADER
    )

    return metaElement?.content || ''
}

/**
 * Sets span as active and executes the callback in span's context.
 *
 * See https://github.com/open-telemetry/opentelemetry-js-api/blob/main/docs/tracing.md#describing-a-span
 */
export function runInSpanContext(span: Span, callback: () => void): void {
    context.with(trace.setSpan(context.active(), span), callback)
}

/**
 * Adds defined time events to span.
 *
 * See https://opentelemetry.io/docs/concepts/signals/traces/#span-events
 */
export function addTimeEventsToSpan<T extends Record<string, TimeInput | undefined>>(
    span: Span,
    timeEvents: T,
    names: Extract<keyof T, string>[]
): void {
    for (const name of names) {
        if (typeof timeEvents[name] !== 'undefined') {
            span.addEvent(name, timeEvents[name])
        }
    }
}
