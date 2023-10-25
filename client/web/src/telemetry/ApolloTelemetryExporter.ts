import { type ApolloClient, gql } from '@apollo/client'
import type { ExportTelemetryEventsResult } from 'src/graphql-operations'

import { TelemetryEventInput, TelemetryExporter } from '@sourcegraph/telemetry'

/**
 * ApolloTelemetryExporter exports events via the new Sourcegraph telemetry
 * framework: https://docs.sourcegraph.com/dev/background-information/telemetry
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
