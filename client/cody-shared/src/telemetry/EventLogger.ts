import * as uuid from 'uuid'

import { version as packageVersion } from '../../package.json'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

interface StorageProvider {
    get(key: string): string | null
    set(key: string, value: string): Promise<void>
}

export class EventLogger {
    private version = packageVersion

    private constructor(private gqlAPIClient: SourcegraphGraphQLAPIClient, private uid: string) {}

    public static async create(
        localStorageService: StorageProvider,
        gqlAPIClient: SourcegraphGraphQLAPIClient
    ): Promise<EventLogger> {
        let anonymousUserID = localStorageService.get(ANONYMOUS_USER_ID_KEY)
        let newInstall = false
        if (!anonymousUserID) {
            newInstall = true
            anonymousUserID = uuid.v4()
            await localStorageService.set(ANONYMOUS_USER_ID_KEY, anonymousUserID)
        }
        const eventLogger = new EventLogger(gqlAPIClient, anonymousUserID)
        if (newInstall) {
            void eventLogger.log('CodyInstalled')
        }
        return eventLogger
    }

    /**
     * @param eventName The ID of the action executed.
     */
    public async log(eventName: string, eventProperties?: any, publicProperties?: any): Promise<void> {
        // Don't log events if the UID has not yet been generated.
        if (this.uid === null) {
            return
        }
        const argument = { ...eventProperties, version: this.version }
        const publicArgument = { ...publicProperties, version: this.version }

        try {
            await this.gqlAPIClient.logEvent({
                event: eventName,
                userCookieID: this.uid,
                source: 'CODY',
                url: '',
                argument: JSON.stringify(argument),
                publicArgument: JSON.stringify(publicArgument),
            })
        } catch (error) {
            console.log(error)
        }
    }
}
