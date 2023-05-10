import cookies from 'js-cookie'
import * as vscode from 'vscode'

import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

import { LocalStorage } from './LocalStorageProvider'

function _getServerEndpointFromConfig(config: vscode.WorkspaceConfiguration): string {
    return config.get<string>('cody.serverEndpoint', '')
}

export const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

const config = vscode.workspace.getConfiguration()
const memento = vscode.workspace.getConfiguration().get<vscode.Memento>('cody.memento')
let localStorage: LocalStorage | undefined
if (memento) {
    localStorage = new LocalStorage(memento)
}

let anonymousUserID = cookies.get(ANONYMOUS_USER_ID_KEY)
if (anonymousUserID === null && localStorage !== undefined) {
    anonymousUserID = localStorage.getAnonymousUserID() as string
}

export class EventLogger {
    private serverEndpoint = _getServerEndpointFromConfig(config)
    private extensionDetails = { ide: 'VSCode', ideExtensionType: 'Cody' }

    private constructor(private gqlAPIClient: SourcegraphGraphQLAPIClient) {}

    public static create(gqlAPIClient: SourcegraphGraphQLAPIClient): EventLogger {
        return new EventLogger(gqlAPIClient)
    }

    /**
     * @param eventName The ID of the action executed.
     */
    public async log(eventName: string, eventProperties?: any, publicProperties?: any): Promise<void> {
        const anonymousUserID =
            cookies.get(ANONYMOUS_USER_ID_KEY) || (localStorage ? localStorage.getAnonymousUserID() : undefined)

        // Don't log events if the UID has not yet been generated.
        if (!anonymousUserID) {
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
                userCookieID: anonymousUserID,
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
