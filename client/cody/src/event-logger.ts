import { Configuration } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

import { LocalStorage } from './command/LocalStorageProvider'
import { sanitizeServerEndpoint } from './sanitize'
import { SecretStorage, getAccessToken } from './secret-storage'

let eventLogger: EventLogger | null = null
let eventServerEndpoint: string

export async function updateEventLogger(
    config: Configuration,
    secretStorage: SecretStorage,
    localStorage: LocalStorage
): Promise<void> {
    const accessToken = await getAccessToken(secretStorage)
    eventServerEndpoint = sanitizeServerEndpoint(config.serverEndpoint)
    const gqlAPIClient = new SourcegraphGraphQLAPIClient(eventServerEndpoint, accessToken, config.customHeaders)
    eventLogger = await EventLogger.create(localStorage, gqlAPIClient)
}

export function logEvent(eventName: string, eventProperties?: any, publicProperties?: any): void {
    if (!eventLogger) {
        return
    }

    let eventPropertiesWithServerEndpoint = { serverEndpoint: eventServerEndpoint }
    let publicPropertiesWithServerEndpoint = { serverEndpoint: eventServerEndpoint }

    if (eventProperties) {
        eventPropertiesWithServerEndpoint = {
            ...eventProperties,
            serverEndpoint: eventServerEndpoint,
        }
    }
    if (publicProperties) {
        publicPropertiesWithServerEndpoint = {
            ...publicProperties,
            serverEndpoint: eventServerEndpoint,
        }
    }
    void eventLogger.log(eventName, eventPropertiesWithServerEndpoint, publicPropertiesWithServerEndpoint)
}
