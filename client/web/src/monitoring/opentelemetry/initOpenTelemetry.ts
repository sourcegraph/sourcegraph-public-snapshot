import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web'
import { ZoneContextManager } from '@opentelemetry/context-zone'

import { Resource } from '@opentelemetry/resources'
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions'

let isInitialized = false
function collectorURL(): string | null {
    // return process.env.NODE_ENV === 'production' ? window.context.opentelemetryCollectorURL : null
    return window.context.opentelemetryCollectorURL
}

export function initOpenTelemetry(): void {
    const url = collectorURL()
    console.log({ opentelemetryURL: url })
    if (url && !isInitialized) {
        isInitialized = true
        const exporter = new OTLPTraceExporter({ url })
        const provider = new WebTracerProvider({
            resource: new Resource({
                [SemanticResourceAttributes.SERVICE_NAME]: 'browser',
            }),
        })
        provider.addSpanProcessor(new BatchSpanProcessor(exporter))
        provider.register({
            contextManager: new ZoneContextManager(),
        })
    }
}
