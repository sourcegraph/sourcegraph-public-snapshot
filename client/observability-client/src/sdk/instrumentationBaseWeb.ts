import { type Instrumentation, type InstrumentationConfig, InstrumentationBase } from '@opentelemetry/instrumentation'

import { createActiveSpan } from './createActiveSpan'
import { createFinishedSpan } from './createFinishedSpan'

/**
 * Base abstract class for instrumenting Sourcegraph OpenTelemetry web plugins.
 *
 * The implementation is based on
 * https://github.com/open-telemetry/opentelemetry-js/tree/main/experimental/packages/opentelemetry-instrumentation
 */
export abstract class InstrumentationBaseWeb extends InstrumentationBase implements Instrumentation {
    protected createActiveSpan = createActiveSpan.bind(this, this.tracer)
    protected createFinishedSpan = createFinishedSpan.bind(this, this.tracer)

    constructor(
        instrumentationName: string,
        instrumentationVersion: string,
        // Do not enable instrumentation by default until `registerInstrumentations` call.
        config: InstrumentationConfig = { enabled: false }
    ) {
        super(instrumentationName, instrumentationVersion, config)
    }

    public init(): void {
        /** noop, the abstract method is defined for overriding modules in node.js */
    }
}
