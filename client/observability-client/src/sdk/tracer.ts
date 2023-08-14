import type { Tracer } from '@opentelemetry/sdk-trace-base'
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web'

import { getFetchInstrumentationStartSpan } from '../instrumentations/fetch'

export class SourcegraphWebTracerProvider extends WebTracerProvider {
    public getTracer(name: string, version?: string, options?: { schemaUrl?: string }): Tracer {
        const tracer = super.getTracer(name, version, options)
        const originalStartSpan = tracer.startSpan.bind(tracer)

        /**
         * Patch `tracer.startSpan` specifically for `@opentelemetry/instrumentation-fetch` because
         * the instrumentation doesn't provide an API for controlling spans context.
         *
         * See https://github.com/open-telemetry/opentelemetry-js/issues/3237
         */
        if (tracer.instrumentationLibrary.name === '@opentelemetry/instrumentation-fetch') {
            tracer.startSpan = getFetchInstrumentationStartSpan(originalStartSpan)
        }

        return tracer
    }
}
