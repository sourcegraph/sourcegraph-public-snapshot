import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

import { version as packageVersion } from '../package.json'

import { LocalStorage } from './services/LocalStorageProvider'

let eventLoggerGQLClient: SourcegraphGraphQLAPIClient
let eventLogger: EventLogger | null = null
let anonymousUserID: string

export async function updateEventLogger(
    config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
    localStorage: LocalStorage
): Promise<void> {
    const status = await localStorage.setAnonymousUserID()
    anonymousUserID = localStorage.getAnonymousUserID() || ''
    if (!eventLoggerGQLClient) {
        eventLoggerGQLClient = new SourcegraphGraphQLAPIClient(config)
        eventLogger = EventLogger.create(eventLoggerGQLClient)
        if (status === 'installed') {
            logEvent('CodyInstalled')
        } else {
            logEvent('CodyVSCodeExtension:CodySavedLogin:executed')
        }
    } else {
        eventLoggerGQLClient.onConfigurationChange(config)
    }
}

/**
 * Logs an event.
 *
 * PRIVACY: Do NOT include any potentially private information in this field.
 * These properties get sent to our analytics tools for Cloud, so must not
 * include private information, such as search queries or repository names.
 *
 * @param eventName The name of the event.
 * @param eventProperties The additional argument information.
 * @param publicProperties Public argument information.
 */
export function logEvent(eventName: string, eventProperties?: any, publicProperties?: any): void {
    if (!eventLogger || !anonymousUserID) {
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
    try {
        eventLogger.log(eventName, anonymousUserID, argument, publicArgument)
    } catch (error) {
        console.error(error)
    }
}
