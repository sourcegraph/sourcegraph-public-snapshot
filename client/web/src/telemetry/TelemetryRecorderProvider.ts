import { type ApolloClient } from '@apollo/client'

import {
    TelemetryRecorderProvider as BaseTelemetryRecorderProvider,
    NoOpTelemetryExporter,
} from '@sourcegraph/telemetry'

import type { BillingCategory, BillingProduct, EventAction, EventFeature, MetadataKey } from '.'
import { ApolloTelemetryExporter } from './ApolloTelemetryExporter'

/**
 * TelemetryRecorderProvider is the default provider implementation.
 */
export class TelemetryRecorderProvider extends BaseTelemetryRecorderProvider<
    EventFeature,
    EventAction,
    MetadataKey,
    BillingCategory,
    BillingProduct
> {
    constructor(client: ApolloClient<object>) {
        super({ client: 'web', clientVersion: 'TODO' }, new ApolloTelemetryExporter(client))
    }
}

class NoOpTelemetryRecorderProvider extends BaseTelemetryRecorderProvider<
    EventFeature,
    EventAction,
    MetadataKey,
    BillingCategory,
    BillingProduct
> {
    constructor() {
        super({ client: '' }, new NoOpTelemetryExporter(), [])
    }
}

export const noOptelemetryRecorderProvider = new NoOpTelemetryRecorderProvider()
export const noOpTelemetryRecorder = new NoOpTelemetryRecorderProvider().getRecorder()
