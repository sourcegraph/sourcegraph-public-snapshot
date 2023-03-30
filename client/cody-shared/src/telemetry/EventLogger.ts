import * as uuid from 'uuid'

import { version as packageVersion } from '../../package.json'
// import { LocalStorage } from './LocalStorageProvider'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

interface StorageProvider {
    get(key: string): string | null
    set(key: string, value: string): Promise<void>
}

export class EventLogger {
    private gqlAPIClient: SourcegraphGraphQLAPIClient
    private uid: string | null = null
    private version = packageVersion
    private localStorageService: StorageProvider
    private newInstall: boolean = false

    constructor(storage: StorageProvider, gqlAPIClient: SourcegraphGraphQLAPIClient) {
        this.localStorageService = storage
        this.gqlAPIClient = gqlAPIClient
        this.initializeLogParameters()
            .then(() => {})
            .catch(() => {})
    }

    private async initializeLogParameters(): Promise<void> {
        let anonymousUserID = this.localStorageService.get(ANONYMOUS_USER_ID_KEY)
        if (!anonymousUserID) {
            anonymousUserID = uuid.v4()
            this.newInstall = true
            await this.localStorageService.set(ANONYMOUS_USER_ID_KEY, anonymousUserID)
        }
        this.uid = anonymousUserID
        if (this.newInstall) {
            this.log('CodyInstalled')
            this.newInstall = false
        }
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @todo Handle arbitrary action IDs.
     *
     * @param eventName The ID of the action executed.
     */
    public async log(eventName: string, eventProperties?: any, publicProperties?: any): Promise<void> {
        // Don't log events if the UID has not yet been generated.
        if (this.uid == null) {
            return
        }
        const argument = { ...eventProperties, version: this.version }
        const publicArgument = { ...publicProperties, version: this.version }

        try {
            await this.gqlAPIClient.logEvent({ name: eventName, userCookieID: this.uid, url: '', argument, publicArgument })
        } catch (error) {
            console.log(error)
        }
    }
}
