import type { ApolloClient } from '@apollo/client'

import type { BillingCategory, BillingProduct } from '@sourcegraph/shared/src/telemetry'
import { sessionTracker } from '@sourcegraph/shared/src/telemetry/web/sessionTracker'
import { userTracker } from '@sourcegraph/shared/src/telemetry/web/userTracker'
import {
    TelemetryRecorderProvider as BaseTelemetryRecorderProvider,
    MarketingTrackingTelemetryProcessor,
    type MarketingTrackingProvider,
    type TelemetryEventMarketingTrackingInput,
} from '@sourcegraph/telemetry'

import { ApolloTelemetryExporter } from './apolloTelemetryExporter'

export function getTelemetrySourceClient(): string {
    if (window.context?.sourcegraphDotComMode) {
        return 'dotcom.web'
    }
    return 'server.web'
}

/**
 * TelemetryRecorderProvider is the default provider implementation for the
 * Sourcegraph web app.
 */
export class TelemetryRecorderProvider extends BaseTelemetryRecorderProvider<BillingCategory, BillingProduct> {
    constructor(
        apolloClient: ApolloClient<object>,
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
                client: getTelemetrySourceClient(),
                clientVersion: window.context.version,
            },
            new ApolloTelemetryExporter(apolloClient),
            [new MarketingTrackingTelemetryProcessor(new TrackingMetadataProvider())],
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

class TrackingMetadataProvider implements MarketingTrackingProvider {
    private user = userTracker
    private session = sessionTracker

    public getMarketingTrackingMetadata(): TelemetryEventMarketingTrackingInput | null {
        if (!window.context?.sourcegraphDotComMode) {
            return null // don't report this data outside of dotcom
        }

        return {
            cohortID: this.user.cohortID,
            deviceSessionID: this.user.deviceSessionID,
            firstSourceURL: this.session.getFirstSourceURL(),
            lastSourceURL: this.session.getLastSourceURL(),
            referrer: this.session.getReferrer(),
            sessionFirstURL: this.session.getSessionFirstURL(),
            sessionReferrer: this.session.getSessionReferrer(),
            url: window.location.href,
        }
    }
}
