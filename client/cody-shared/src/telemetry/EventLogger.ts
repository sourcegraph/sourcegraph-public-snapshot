import * as uuid from 'uuid'

import { version as packageVersion } from '../../package.json'
import { ANONYMOUS_USER_ID_KEY, LocalStorageService } from '../localStorage'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

export class EventLogger {
    private apiClient: SourcegraphGraphQLAPIClient
    private uid: string | null = null
    private version = packageVersion
    private localStorageService: LocalStorageService
    private newInstall: boolean = false

    constructor(storage: LocalStorageService, apiClient: SourcegraphGraphQLAPIClient) {
        this.localStorageService = storage
        this.apiClient = apiClient
        this.initializeLogParameters()
            .then(() => {})
            .catch(() => {})
    }

    private async initializeLogParameters(): Promise<void> {
        let anonymousUserID = this.localStorageService.getValue(ANONYMOUS_USER_ID_KEY)
        if (!anonymousUserID) {
            anonymousUserID = uuid.v4()
            this.newInstall = true
            await this.localStorageService.setValue(ANONYMOUS_USER_ID_KEY, anonymousUserID)
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
        await this.apiClient.logEvent({name: eventName, userCookieID: this.uid, url: "", argument, publicArgument})
    }
}
