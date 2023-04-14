import { Configuration } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

import { LocalStorage } from './command/LocalStorageProvider'
import { sanitizeServerEndpoint } from './sanitize'
import { SecretStorage, getAccessToken } from './secret-storage'

let eventLogger: EventLogger | null = null

export async function updateEventLogger(
    config: Configuration,
    secretStorage: SecretStorage,
    localStorage: LocalStorage
): Promise<void> {
    const accessToken = await getAccessToken(secretStorage)
    const gqlAPIClient = new SourcegraphGraphQLAPIClient(
        sanitizeServerEndpoint(config.serverEndpoint),
        accessToken,
        config.customHeaders
    )
    eventLogger = await EventLogger.create(localStorage, gqlAPIClient)
}

export function logEvent(eventName: string, eventProperties?: any, publicProperties?: any): void {
    if (!eventLogger) {
        return
    }
    void eventLogger.log(eventName, eventProperties, publicProperties)
}
