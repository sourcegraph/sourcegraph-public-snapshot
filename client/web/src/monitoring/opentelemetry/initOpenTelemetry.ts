// Order is important here.
// Don't remove the empty lines between these imports.
import './initZones'

import { ZoneContextManager } from '@opentelemetry/context-zone'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { InstrumentationOption, registerInstrumentations } from '@opentelemetry/instrumentation'
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch'
import { Resource } from '@opentelemetry/resources'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web'
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'
import isAbsoluteUrl from 'is-absolute-url'

import {
    ConsoleBatchSpanExporter,
    WindowLoadInstrumentation,
    HistoryInstrumentation,
    SharedSpanStoreProcessor,
    ClientAttributesSpanProcessor,
} from '@sourcegraph/observability-client'

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
        const provider = new WebTracerProvider({
            resource: new Resource({
                [SemanticResourceAttributes.SERVICE_NAME]: 'web-app',
            }),
        })

        const url = isAbsoluteUrl(openTelemetry.endpoint)
            ? openTelemetry.endpoint
            : new URL(openTelemetry.endpoint, externalURL).toString()

        // As per spec non-signal-specific configuration should have signal-specific paths appended.
        // https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#endpoint-urls-for-otlphttp
        const collectorExporter = new OTLPTraceExporter({ url: new URL('/v1/traces', url).toString() })

        provider.addSpanProcessor(new BatchSpanProcessor(collectorExporter))
        provider.addSpanProcessor(new SharedSpanStoreProcessor())
        provider.addSpanProcessor(new ClientAttributesSpanProcessor(window.context.version))

        // Enable the console exporter only in the development environment.
        if (process.env.NODE_ENV === 'development') {
            const consoleExporter = new ConsoleBatchSpanExporter()
            provider.addSpanProcessor(new BatchSpanProcessor(consoleExporter))
        }

        provider.register({
            contextManager: new ZoneContextManager(),
        })

        registerInstrumentations({
            // Type-casting is required since the `FetchInstrumentation` is wrongly typed internally as `node.js` instrumentation.
            instrumentations: [
                (new FetchInstrumentation() as unknown) as InstrumentationOption,
                new WindowLoadInstrumentation(),
                new HistoryInstrumentation(),
            ],
        })
    }
}
