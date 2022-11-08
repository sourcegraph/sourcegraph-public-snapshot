import { trace, context, Context, Span } from '@opentelemetry/api'
import { ReadableSpan } from '@opentelemetry/sdk-trace-base'

export enum SharedSpanName {
    PageView = 'PageView',
    WindowLoad = 'WindowLoad',
}

type SharedSpanNames = keyof typeof SharedSpanName

/**
 * Used to store recent navigation spans to group other types of spans
 * under them, which helps analyze data in Honeycomb.
 *
 * Question: Why having a separate store instead of sharing via context manager?
 *
 * The OpenTelemetry context is immutable and can only be passed down
 * with a callback. If there's no way to wrap the function execution into a
 * parent span via callback we need to implement another sharing mechanism like
 * a store. This issue is raised in the OpenTelemetry repo and currently does not
 * have a recommended solution.
 *
 * Example issues:
 * 1. https://github.com/open-telemetry/opentelemetry-js-contrib/issues/995#issuecomment-1138367723
 * 2. https://github.com/open-telemetry/opentelemetry-js-contrib/issues/732
 */
class SharedSpanStore {
    private spanMap: { [key in SharedSpanNames]?: { context: Context; span: ReadableSpan } } = {}

    public set(spanName: SharedSpanName, span: ReadableSpan): void {
        this.spanMap[spanName] = {
            span,
            context: trace.setSpan(context.active(), (span as unknown) as Span),
        }
    }

    /**
     * Get context created by the most recent navigation span.
     * Context created by either `PageView` or `WindowLoad` spans.
     */
    public getRootNavigationContext(): Context | undefined {
        return (this.spanMap.PageView || this.spanMap.WindowLoad)?.context
    }

    /**
     * Get the most recent navigation span: either `PageView` or `WindowLoad` spans.
     */
    public getRootNavigationSpan(): ReadableSpan | undefined {
        return (this.spanMap.PageView || this.spanMap.WindowLoad)?.span
    }
}

export const sharedSpanStore = new SharedSpanStore()
