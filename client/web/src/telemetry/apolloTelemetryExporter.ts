import { type ApolloClient, gql } from '@apollo/client'

import type { TelemetryEventInput, TelemetryExporter } from '@sourcegraph/telemetry'

import type { ExportTelemetryEventsResult } from '../graphql-operations'

/**
 * ApolloTelemetryExporter exports events via the new Sourcegraph telemetry
 * framework: https://docs-legacy.sourcegraph.com/dev/background-information/telemetry
 */
export class ApolloTelemetryExporter implements TelemetryExporter {
    constructor(private client: ApolloClient<object>) {}

    public async exportEvents(events: TelemetryEventInput[]): Promise<void> {
        await this.client.mutate<ExportTelemetryEventsResult>({
            mutation: gql`
                mutation ExportTelemetryEvents($events: [TelemetryEventInput!]!) {
                    telemetry {
                        recordEvents(events: $events) {
                            alwaysNil
                        }
                    }
                }
            `,
            variables: { events },
        })
    }
}
