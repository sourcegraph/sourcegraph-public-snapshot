import { Memento } from 'vscode'

import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

let eventLoggerGQLClient: SourcegraphGraphQLAPIClient
let eventLogger: EventLogger | null = null

export class LocalStorage {
    constructor(private storage: Memento) {}

    public get(key: string): string | null {
        return this.storage.get(key, null)
    }

    public async set(key: string, value: string): Promise<void> {
        try {
            await this.storage.update(key, value)
        } catch (error) {
            console.error(error)
        }
    }
}

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
        console.log('Event logger not initialized')
        return
    }
    void eventLogger.log(eventName, eventProperties, publicProperties)
}
