import { type Span, trace, context, type Context } from '@opentelemetry/api'
import { otperformance } from '@opentelemetry/core'
import { PerformanceTimingNames } from '@opentelemetry/sdk-trace-web'
import { SemanticAttributes } from '@opentelemetry/semantic-conventions'

import {
    addTimeEventsToSpan,
    addSpanPerformancePaintEvents,
    performanceNavigationTimingToEntries,
    InstrumentationBaseWeb,
    SharedSpanName,
} from '../sdk'

import { WindowLoadSpanName } from './constants'

/**
 * Auto instrumentation of the window load event based on Web Performance API `navigation` entries.
 *
 * 1. Listens to the first performance `navigation` entries update.
 * 2. Creates the `WINDOW_LOAD` span capturing timings from `FETCH_START` till `LOAD_EVENT_END`.
 * 3. Adds performance `navigation` events to the `WINDOW_LOAD` span.
 * 4. Adds performance `paint` events to the `WINDOW_LOAD` span.
 * 5. Creates nested spans for all resources loaded before the `LOAD_EVENT_END`.
 *
 * See Navigation Timing API documentation
 * https://developer.mozilla.org/en-US/docs/Web/API/Navigation_timing_API
 *
 * See Navigation Timing spec processing model:
 * https://www.w3.org/TR/navigation-timing-2/#processing-model
 *
 * See Resource Timing spec processing model:
 * https://www.w3.org/TR/resource-timing-2/#attribute-descriptions
 *
 * Based on the OpenTelemetry Instrumentation Document Load
 * https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/plugins/web/opentelemetry-instrumentation-document-load
 *
 * The implementation is forked because of various issues blocking
 * the integration with other auto instrumentations.
 *
 * RUM integration: https://github.com/open-telemetry/opentelemetry-js-contrib/issues/732
 * Fetch integration: https://github.com/open-telemetry/opentelemetry-js-contrib/issues/995
 */
export class WindowLoadInstrumentation extends InstrumentationBaseWeb {
    public static instrumentationName = '@sourcegraph/instrumentation-window-load'
    public static version = '0.1'

    private rootContext: Context | undefined

    constructor() {
        super(WindowLoadInstrumentation.instrumentationName, WindowLoadInstrumentation.version)
    }

    private addResourcesSpans(parentSpan: Span): void {
        // Casting is required until this issue is resolved: https://github.com/microsoft/TypeScript/issues/33866
        for (const resource of otperformance.getEntriesByType('resource') as PerformanceResourceTiming[]) {
            this.createFinishedSpan({
                name: WindowLoadSpanName.ResourceFetch,
                startTime: resource[PerformanceTimingNames.FETCH_START],
                endTime: resource[PerformanceTimingNames.RESPONSE_END],
                parentSpan,
                networkEvents: resource,
                attributes: {
                    [SemanticAttributes.HTTP_URL]: resource.name,
                },
            })
        }
    }

    private collectPerformance(): void {
        if (!this.rootContext) {
            throw new Error('The `WindowLoad` context should be created. Check `createWindowLoadContext` method.')
        }

        const rootSpan = trace.getSpan(this.rootContext)

        if (!rootSpan) {
            throw new Error('The `WindowLoad` span should be created. Check `createWindowLoadContext` method.')
        }

        const entries = performanceNavigationTimingToEntries()

        context.with(this.rootContext, () => {
            addTimeEventsToSpan(rootSpan, entries, [
                PerformanceTimingNames.FETCH_START,
                PerformanceTimingNames.UNLOAD_EVENT_START,
                PerformanceTimingNames.UNLOAD_EVENT_END,
                PerformanceTimingNames.DOM_INTERACTIVE,
                PerformanceTimingNames.DOM_CONTENT_LOADED_EVENT_START,
                PerformanceTimingNames.DOM_CONTENT_LOADED_EVENT_END,
                PerformanceTimingNames.DOM_COMPLETE,
                PerformanceTimingNames.LOAD_EVENT_START,
                PerformanceTimingNames.LOAD_EVENT_END,
            ])

            addSpanPerformancePaintEvents(rootSpan)

            this.addResourcesSpans(rootSpan)
            this.createFinishedSpan({
                name: WindowLoadSpanName.DocumentFetch,
                startTime: entries[PerformanceTimingNames.FETCH_START],
                endTime: entries[PerformanceTimingNames.RESPONSE_END],
                networkEvents: entries,
                parentSpan: rootSpan,
            })

            rootSpan.end(entries[PerformanceTimingNames.LOAD_EVENT_END])
        })
    }

    public createWindowLoadContext(): void {
        const entries = performanceNavigationTimingToEntries()

        const span = this.tracer.startSpan(SharedSpanName.WindowLoad, {
            startTime: entries[PerformanceTimingNames.FETCH_START],
        })

        this.rootContext = trace.setSpan(context.active(), span)
    }

    private onDocumentLoaded = (): void => {
        // Timeout is needed as load event doesn't have yet the performance metrics for loadEnd.
        setTimeout(() => {
            this.collectPerformance()
        })
    }

    public enable(): void {
        // Create the WindowLoad span on instrumentation init to make it possible
        // to link spans from other instrumentation to it, which occur before the WindowLoad event
        // E.g., `@opentelemetry/instrumentation-fetch` spans.
        this.createWindowLoadContext()

        if (window.document.readyState === 'complete') {
            return this.collectPerformance()
        }

        window.addEventListener('load', this.onDocumentLoaded)
    }

    public disable(): void {
        window.removeEventListener('load', this.onDocumentLoaded)
    }
}
