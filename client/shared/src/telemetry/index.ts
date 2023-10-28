import {
    TelemetryRecorderProvider as BaseTelemetryRecorderProvider,
    NoOpTelemetryExporter,
    type TelemetryProcessor,
    CallbackTelemetryProcessor,
} from '@sourcegraph/telemetry'

/**
 * TelemetryRecorderProvider type used in Sourcegraph clients.
 */
export type TelemetryRecorderProvider = typeof noOptelemetryRecorderProvider

/**
 * TelemetryRecorder type used in Sourcegraph clients.
 */
export type TelemetryRecorder = typeof noOpTelemetryRecorder

/**
 * Events accept billing metadata for ease of categorization in analytics
 * pipelines - this type enumerates known categories.
 */
export type BillingCategory = 'exampleBillingCategory'

/**
 * Events accept billing metadata for ease of categorization in analytics
 * pipelines - this type enumerates known products.
 */
export type BillingProduct = 'exampleBillingProduct'

/**
 * Props interface that can be extended by React components depending on the
 * new telemetry framework: https://docs.sourcegraph.com/dev/background-information/telemetry
 * These properties are part of {@link PlatformContext}.
 */
export interface TelemetryV2Props {
    /**
     * Telemetry recorder for the new telemetry framework, superseding
     * 'telemetryService' and 'logEvent' variants. Learn more here:
     * https://docs.sourcegraph.com/dev/background-information/telemetry
     *
     * It is backed by a '@sourcegraph/telemetry' implementation.
     */
    telemetryRecorder: TelemetryRecorder
}

export class NoOpTelemetryRecorderProvider extends BaseTelemetryRecorderProvider<BillingCategory, BillingProduct> {
    constructor(opts?: { errorOnRecord?: boolean }) {
        const processors: TelemetryProcessor[] = []
        if (opts?.errorOnRecord) {
            processors.push(
                new CallbackTelemetryProcessor(() => {
                    throw new Error('telemetry: unexpected use of no-op telemetry recorder')
                })
            )
        }
        super({ client: '' }, new NoOpTelemetryExporter(), processors)
    }
}

export const noOptelemetryRecorderProvider = new NoOpTelemetryRecorderProvider()
export const noOpTelemetryRecorder = noOptelemetryRecorderProvider.getRecorder()
