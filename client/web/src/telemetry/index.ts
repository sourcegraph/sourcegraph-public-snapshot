import type { noOpTelemetryRecorder, noOptelemetryRecorderProvider } from './TelemetryRecorderProvider'

/**
 * TelemetryRecorderProvider type used in Sourcegraph web.
 */
export type TelemetryRecorderProvider = typeof noOptelemetryRecorderProvider

/**
 * TelemetryRecorder type used in Sourcegraph web.
 */
export type TelemetryRecorder = typeof noOpTelemetryRecorder

/**
 * Features indicate the functionality being tracked.
 *
 * All Cody features must start with `cody.`, for example `cody.myFeature`.
 */
export type EventFeature = 'exampleFeature'

/**
 * Actions should denote a generic action within the scope of a feature. Where
 * possible, reuse an existing action.
 */
export type EventAction = 'succeeded' | 'failed'

/**
 * MetadataKey is an allowlist of keys for the safe-for-export metadata parameter.
 * Where possible, reuse an existing key.
 */
export type MetadataKey = 'durationMs'

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
