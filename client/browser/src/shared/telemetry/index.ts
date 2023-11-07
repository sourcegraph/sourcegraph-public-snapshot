import type { ApolloClient } from '@apollo/client'

import type { BillingCategory, BillingProduct } from '@sourcegraph/shared/src/telemetry'
import { ApolloTelemetryExporter } from '@sourcegraph/shared/src/telemetry/apolloTelemetryExporter'
import {
    TelemetryRecorderProvider as BaseTelemetryRecorderProvider,
    MarketingTrackingTelemetryProcessor,
    type MarketingTrackingProvider,
    type TelemetryEventMarketingTrackingInput,
} from '@sourcegraph/telemetry'

import { getExtensionVersion, getPlatformName } from '../util/context'

/**
 * TelemetryRecorderProvider is the default provider implementation for the
 * Sourcegraph web app.
 */
export class TelemetryRecorderProvider extends BaseTelemetryRecorderProvider<BillingCategory, BillingProduct> {
    constructor(
        apolloClient: Pick<ApolloClient<object>, 'mutate'>,
        options: {
            /**
             * Enables buffering of events for export. Only enable if there is a
             * reliable unsubscribe mechanism available.
             */
            enableBuffering: boolean
        }
    ) {
        super(
            {
                client: getPlatformName(),
                clientVersion: getExtensionVersion(),
            },
            new ApolloTelemetryExporter(apolloClient),
            [],
            {
                /**
                 * Use buffer time of 100ms - some existing buffering uses
                 * 1000ms, but we use a more conservative value.
                 */
                bufferTimeMs: options.enableBuffering ? 100 : 0,
                bufferMaxSize: 10,
                errorHandler: error => {
                    throw new Error(error)
                },
            }
        )
    }
}
