import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'
import { version as packageVersion } from '../package.json'

import { LocalStorage } from './command/LocalStorageProvider'

let eventLoggerGQLClient: SourcegraphGraphQLAPIClient
let eventLogger: EventLogger | null = null
let version = packageVersion

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

    const argument = {
        ...eventProperties,
        version: this.version,
    }

    const publicArgument = {
        ...publicProperties,
        version: this.version,
    }

    void eventLogger.log(eventName, argument, publicArgument)
}
