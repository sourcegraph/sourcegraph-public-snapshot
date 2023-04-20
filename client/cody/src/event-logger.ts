import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

import { LocalStorage } from './command/LocalStorageProvider'

let eventLoggerGQLClient: SourcegraphGraphQLAPIClient
let eventLogger: EventLogger | null = null

export async function updateEventLogger(
    config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
    localStorage: LocalStorage
): Promise<void> {
    if (!eventLoggerGQLClient) {
        eventLoggerGQLClient = new SourcegraphGraphQLAPIClient(config)
        eventLogger = await EventLogger.create(localStorage, eventLoggerGQLClient)
    } else {
        eventLoggerGQLClient.onConfigurationChange(config)
    }
}

export function logEvent(eventName: string, eventProperties?: any, publicProperties?: any): void {
    if (!eventLogger) {
        return
    }
    void eventLogger.log(eventName, eventProperties, publicProperties)
}
