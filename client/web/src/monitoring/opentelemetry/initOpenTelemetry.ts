import { ZoneContextManager } from '@opentelemetry/context-zone'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { registerInstrumentations } from '@opentelemetry/instrumentation'
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch'
import { Resource } from '@opentelemetry/resources'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web'
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'
import isAbsoluteUrl from 'is-absolute-url'

export function initOpenTelemetry(): void {
    if (process.env.NODE_ENV === 'production' || process.env.ENABLE_MONITORING) {
        const provider = new WebTracerProvider({
            resource: new Resource({
                [SemanticResourceAttributes.SERVICE_NAME]: 'web-app',
            }),
        })

        const { openTelemetryTracing, externalURL } = window.context

        if (openTelemetryTracing) {
            const url = isAbsoluteUrl(openTelemetryTracing.collectorURL)
                ? openTelemetryTracing.collectorURL
                : `${externalURL}/${openTelemetryTracing.collectorURL}`

            const exporter = new OTLPTraceExporter({ url })
            const spanProcessor = new BatchSpanProcessor(exporter)

            provider.addSpanProcessor(spanProcessor)
        }

        provider.register({
            contextManager: new ZoneContextManager(),
        })

        registerInstrumentations({
            instrumentations: [new FetchInstrumentation()],
        })
    }
}
