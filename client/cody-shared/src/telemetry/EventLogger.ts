import * as uuid from 'uuid'
import * as vscode from 'vscode'

import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

function _getServerEndpointFromConfig(config: vscode.WorkspaceConfiguration): string {
    return config.get<string>('cody.serverEndpoint', '')
}

const config = vscode.workspace.getConfiguration()

const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

interface StorageProvider {
    get(key: string): string | null
    set(key: string, value: string): Promise<void>
}

export class EventLogger {
    private serverEndpoint = _getServerEndpointFromConfig(config)
    private extensionDetails = { ide: 'VSCode', ideExtensionType: 'Cody' }

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
            await eventLogger.log('CodyInstalled')
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
        const argument = {
            ...eventProperties,
            serverEndpoint: this.serverEndpoint,
            extensionDetails: this.extensionDetails,
        }
        const publicArgument = {
            ...publicProperties,
            serverEndpoint: this.serverEndpoint,
            extensionDetails: this.extensionDetails,
        }

        try {
            await this.gqlAPIClient.logEvent({
                event: eventName,
                userCookieID: this.uid,
                source: 'IDEEXTENSION',
                url: '',
                argument: JSON.stringify(argument),
                publicArgument: JSON.stringify(publicArgument),
            })
        } catch (error) {
            console.log(error)
        }
    }
}
