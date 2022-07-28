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
    const { openTelemetry, externalURL } = window.context

    if (openTelemetry && (process.env.NODE_ENV === 'production' || process.env.ENABLE_OPEN_TELEMETRY)) {
        const provider = new WebTracerProvider({
            resource: new Resource({
                [SemanticResourceAttributes.SERVICE_NAME]: 'web-app',
            }),
        })

        if (openTelemetry) {
            const url = isAbsoluteUrl(openTelemetry.endpoint)
                ? openTelemetry.endpoint
                : `${externalURL}/${openTelemetry.endpoint}`

            // As per spec non-signal-specific configuration should have signal-specific paths appended.
            // https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#endpoint-urls-for-otlphttp
            const exporter = new OTLPTraceExporter({ url: url + '/v1/traces' })
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
