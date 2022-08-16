import { Span, trace, context, Context } from '@opentelemetry/api'

export enum SharedSpanName {
    PageView = 'PageView',
    WindowLoad = 'WindowLoad',
}

type SharedSpanNames = keyof typeof SharedSpanName

/**
 * Used to store recent navigation spans to group other types of spans
 * under them, which helps analyze data in Honeycomb.
 *
 * Shared spans are set via the `SharedSpanStoreProcessor`.
 */
class SharedSpanStore {
    private spanMap: { [key in SharedSpanNames]?: Context } = {}

    public set(spanName: SharedSpanName, span: Span): void {
        this.spanMap[spanName] = trace.setSpan(context.active(), span)
    }

    /**
     * Get context created by the most recent navigation span.
     * Context created by either `PageView` or `WindowLoad` spans.
     */
    public getRootNavigationContext(): Context | undefined {
        return this.spanMap.PageView || this.spanMap.WindowLoad
    }

    /**
     * Get the most recent navigation span: either `PageView` or `WindowLoad` spans.
     */
    public getRootNavigationSpan(): Span | undefined {
        const navigationContext = this.getRootNavigationContext()

        if (navigationContext) {
            return trace.getSpan(navigationContext)
        }

        return undefined
    }
}

export const sharedSpanStore = new SharedSpanStore()
