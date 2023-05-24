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
            await logEvent('CodyInstalled')
        } else {
            await logEvent('CodyVSCodeExtension:CodySavedLogin:executed')
        }
    } else {
        eventLoggerGQLClient.onConfigurationChange(config)
    }
}

export async function logEvent(eventName: string, eventProperties?: any, publicProperties?: any): Promise<void> {
    if (!eventLogger || !getAnonymousUserID()) {
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
    await eventLogger.log(eventName, anonymousUserID, argument, publicArgument)
}

export async function logCodyInstalled(): Promise<void> {
    if (!eventLogger || !anonymousUserID) {
        return
    }
    await eventLogger.log('CodyInstalled', anonymousUserID)
}

function getAnonymousUserID(): string {
    if (!anonymousUserID) {
        anonymousUserID = localStorage.getAnonymousUserID()
    }
    return anonymousUserID
}
