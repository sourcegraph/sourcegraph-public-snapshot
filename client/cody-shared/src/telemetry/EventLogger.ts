import * as vscode from 'vscode'

import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

function _getServerEndpointFromConfig(config: vscode.WorkspaceConfiguration): string {
    return config.get<string>('cody.serverEndpoint', '')
}

const config = vscode.workspace.getConfiguration()

export const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

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
        let anonymousUserID = localStorage.get(ANONYMOUS_USER_ID_KEY)

        // Don't log events if the UID has not yet been generated.
        if (anonymousUserID === null) {
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
