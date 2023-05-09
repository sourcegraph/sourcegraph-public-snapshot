import * as uuid from 'uuid'

import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger, ANONYMOUS_USER_ID_KEY } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

import { version as packageVersion } from '../package.json'

import { LocalStorage } from './command/LocalStorageProvider'

let eventLoggerGQLClient: SourcegraphGraphQLAPIClient
let eventLogger: EventLogger | null = null

export async function updateEventLogger(
    config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
    localStorage: LocalStorage
): Promise<void> {
    const anonymousUserID = localStorage.get(ANONYMOUS_USER_ID_KEY)
    if (!anonymousUserID) {
        const anonymousUserID = uuid.v4()
        await localStorage.set(ANONYMOUS_USER_ID_KEY, anonymousUserID)
    }
    if (!eventLoggerGQLClient) {
        eventLoggerGQLClient = new SourcegraphGraphQLAPIClient(config)
        eventLogger = await EventLogger.create(eventLoggerGQLClient)
        await logCodyInstalled()
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
        version: packageVersion,
    }

    const publicArgument = {
        ...publicProperties,
        version: packageVersion,
    }

    void eventLogger.log(eventName, argument, publicArgument)
}

export async function logCodyInstalled(): Promise<void> {
    if (!eventLogger) {
        return
    }
    await eventLogger.log('CodyInstalled')
}
