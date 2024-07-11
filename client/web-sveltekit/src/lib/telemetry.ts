import { gql } from '@urql/core'

import type { BillingCategory, BillingProduct } from '@sourcegraph/shared/src/telemetry'
import { sessionTracker } from '@sourcegraph/shared/src/telemetry/web/sessionTracker'
import { userTracker } from '@sourcegraph/shared/src/telemetry/web/userTracker'
import {
    type MarketingTrackingProvider,
    type TelemetryEventMarketingTrackingInput,
    type TelemetryEventInput,
    type TelemetryExporter,
    MarketingTrackingTelemetryProcessor,
    TelemetryRecorderProvider as BaseTelemetryRecorderProvider,
} from '@sourcegraph/telemetry'

import type { GraphQLClient } from '$lib/graphql'
import { getGraphQLClient } from '$lib/graphql'
import type { ExportTelemetryEventsResult } from '$lib/graphql-operations'

/**
 * TelemetryRecorderProvider is the default provider implementation for the
 * Sourcegraph web app.
 */
export class TelemetryRecorderProvider extends BaseTelemetryRecorderProvider<BillingCategory, BillingProduct> {
    constructor(
        graphQlClient: GraphQLClient,
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
            new GraphQlTelemetryExporter(graphQlClient),
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

function getTelemetrySourceClient(): string {
    if (window.context?.sourcegraphDotComMode) {
        return 'dotcom.svelte-web'
    }
    return 'server.svelte-web'
}

/**
 * ApolloTelemetryExporter exports events via the new Sourcegraph telemetry
 * framework: https://docs-legacy.sourcegraph.com/dev/background-information/telemetry
 */
export class GraphQlTelemetryExporter implements TelemetryExporter {
    constructor(private client: GraphQLClient) {}

    public async exportEvents(events: TelemetryEventInput[]): Promise<void> {
        await this.client.mutation<ExportTelemetryEventsResult>(
            gql`
                mutation ExportTelemetryEvents($events: [TelemetryEventInput!]!) {
                    telemetry {
                        recordEvents(events: $events) {
                            alwaysNil
                        }
                    }
                }
            `,
            { events }
        )
    }
}

class TrackingMetadataProvider implements MarketingTrackingProvider {
    private user = userTracker
    private session = sessionTracker

    public getMarketingTrackingMetadata(): TelemetryEventMarketingTrackingInput | null {
        if (!window.context?.sourcegraphDotComMode) {
            return null // don't report this data outside dotcom
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

export const TELEMETRY_PROVIDER = new TelemetryRecorderProvider(getGraphQLClient(), { enableBuffering: true })
export const TELEMETRY_RECORDER = TELEMETRY_PROVIDER.getRecorder()
