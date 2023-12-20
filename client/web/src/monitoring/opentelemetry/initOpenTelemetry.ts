// Order is important here.
// Don't remove the empty lines between these imports.
import './initZones'

import type { ZoneContextManager } from '@opentelemetry/context-zone'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { type InstrumentationOption, registerInstrumentations } from '@opentelemetry/instrumentation'
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch'
import { Resource } from '@opentelemetry/resources'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'

import {
    ConsoleBatchSpanExporter,
    WindowLoadInstrumentation,
    HistoryInstrumentation,
    SharedSpanStoreProcessor,
    ClientAttributesSpanProcessor,
    getTracingURL,
    SourcegraphWebTracerProvider,
} from '@sourcegraph/observability-client'

import { PageRoutes } from '../../routes.constants'

export function initOpenTelemetry(): void {
    const { openTelemetry, externalURL } = window.context

    /**
     * OpenTelemetry is enabled only if
     * 1. The backend passthrough endpoint is configured in the site configuration.
     * 2. The application is running in the `production` environment or `ENABLE_OPEN_TELEMETRY` is true.
     *
     * The `ENABLE_OPEN_TELEMETRY` env variable is primarily used for local development
     * because client-side OpenTelemetry is not enabled by default yet.
     *
     */
    if (openTelemetry?.endpoint && (process.env.NODE_ENV === 'production' || process.env.ENABLE_OPEN_TELEMETRY)) {
        const provider = new SourcegraphWebTracerProvider({
            resource: new Resource({
                [SemanticResourceAttributes.SERVICE_NAME]: 'web-app',
            }),
        })

        const collectorExporter = new OTLPTraceExporter({ url: getTracingURL(openTelemetry.endpoint, externalURL) })

        // Span processors are executed in the order they were added to the provider.
        provider.addSpanProcessor(new BatchSpanProcessor(collectorExporter))
        provider.addSpanProcessor(new ClientAttributesSpanProcessor(window.context.version))
        provider.addSpanProcessor(new SharedSpanStoreProcessor())

        // Enable the console exporter only in the development environment.
        if (process.env.NODE_ENV === 'development') {
            const consoleExporter = new ConsoleBatchSpanExporter()
            provider.addSpanProcessor(new BatchSpanProcessor(consoleExporter))
        }

        /**
         * This import enables zone.js which patches global web API modules.
         * You can find a list of the patched modules here:
         * https://github.com/angular/angular/blob/main/packages/zone.js/MODULE.md
         *
         * It's added with the `require` statement to avoid polluting stack traces in
         * the development environment when OpenTelemetry is disabled.
         */
        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
        const ZoneContextManager = require('@opentelemetry/context-zone').ZoneContextManager

        provider.register({
            contextManager: new ZoneContextManager() as ZoneContextManager,
        })

        registerInstrumentations({
            // Type-casting is required since the `FetchInstrumentation` is wrongly typed internally as `node.js` instrumentation.
            instrumentations: [
                new FetchInstrumentation({
                    // Ignore adding network events as span events to reduce the volume of events
                    // sent to OpenTelemetry backends such as Honeycomb.
                    ignoreNetworkEvents: true,
                }) as unknown as InstrumentationOption,
                new WindowLoadInstrumentation(),
                new HistoryInstrumentation({
                    shouldCreatePageViewOnLocationChange: prevLocationInfo => {
                        /**
                         * Start a new PageView trace on `location.search` change only
                         * for some pages to avoid spam.
                         */
                        if (location.pathname.endsWith(PageRoutes.Search)) {
                            return (
                                prevLocationInfo.pathname !== location.pathname ||
                                prevLocationInfo.search !== location.search
                            )
                        }

                        return prevLocationInfo.pathname !== location.pathname
                    },
                }),
            ],
        })
    }
}
