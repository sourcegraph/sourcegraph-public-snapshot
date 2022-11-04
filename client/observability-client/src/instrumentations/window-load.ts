import { propagation, Span, ROOT_CONTEXT } from '@opentelemetry/api'
import { otperformance } from '@opentelemetry/core'
import { PerformanceTimingNames } from '@opentelemetry/sdk-trace-web'
import { SemanticAttributes } from '@opentelemetry/semantic-conventions'

import {
    addTimeEventsToSpan,
    addSpanPerformancePaintEvents,
    performanceNavigationTimingToEntries,
    getServerSideTraceParent,
    ActiveSpanConfig,
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

    // The PerformanceObserver listener is executed once for the first `navigation` list update.
    private observer = new PerformanceObserver((_list, observer) => {
        this.collectPerformance()
        observer.disconnect()
    })

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
        const entries = performanceNavigationTimingToEntries()
        const rootContext = propagation.extract(ROOT_CONTEXT, { traceparent: getServerSideTraceParent() })

        const rootSpanConfig: ActiveSpanConfig = {
            name: SharedSpanName.WindowLoad,
            startTime: entries[PerformanceTimingNames.FETCH_START],
            context: rootContext,
            attributes: {
                [SemanticAttributes.HTTP_URL]: location.href,
                [SemanticAttributes.HTTP_USER_AGENT]: navigator.userAgent,
            },
        }

        this.createActiveSpan(rootSpanConfig, rootSpan => {
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

    public enable(): void {
        if (window.document.readyState === 'complete') {
            return this.collectPerformance()
        }

        this.observer.observe({ type: 'navigation' })
    }

    public disable(): void {
        this.observer.disconnect()
    }
}
