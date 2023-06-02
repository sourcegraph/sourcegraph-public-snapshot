import { get } from 'lodash'
import * as vscode from 'vscode'

import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

function getServerEndpointFromConfig(config: vscode.WorkspaceConfiguration): string {
    return config.get<string>('cody.serverEndpoint', '')
}

function getUseContextFromConfig(config: vscode.WorkspaceConfiguration): string {
    return config.get<string>('cody.useContext', '')
}

function getchatPredictionsFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    return config.get<boolean>('cody.experimental.chatPredictions', false)
}

function getinlineFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    return config.get<boolean>('cody.experimental.inline', false)
}

function getnonStopFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    return config.get<boolean>('cody.experimental.nonStop', false)
}

function getsuggestionsFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    return config.get<boolean>('cody.experimental.suggestions', false)
}

const config = vscode.workspace.getConfiguration()

export class EventLogger {
    private serverEndpoint = getServerEndpointFromConfig(config)
    private extensionDetails = { ide: 'VSCode', ideExtensionType: 'Cody' }
    private configurationDetails = {
        contextSelection: getUseContextFromConfig(config),
        chatPredictions: getchatPredictionsFromConfig(config),
        inline: getinlineFromConfig(config),
        nonStop: getnonStopFromConfig(config),
        suggestions: getsuggestionsFromConfig(config),
    }

    private constructor(private gqlAPIClient: SourcegraphGraphQLAPIClient) {}

    public static create(gqlAPIClient: SourcegraphGraphQLAPIClient): EventLogger {
        return new EventLogger(gqlAPIClient)
    }

    /**
     * Logs an event.
     *
     * PRIVACY: Do NOT include any potentially private information in this
     * field. These properties get sent to our analytics tools for Cloud, so
     * must not include private information, such as search queries or
     * repository names.
     *
     * @param eventName The name of the event.
     * @param anonymousUserID The randomly generated unique user ID.
     * @param eventProperties The additional argument information.
     * @param publicProperties Public argument information.
     */
    public log(eventName: string, anonymousUserID: string, eventProperties?: any, publicProperties?: any): void {
        const argument = {
            ...eventProperties,
            serverEndpoint: this.serverEndpoint,
            extensionDetails: this.extensionDetails,
            configurationDetails: this.configurationDetails,
        }
        const publicArgument = {
            ...publicProperties,
            serverEndpoint: this.serverEndpoint,
            extensionDetails: this.extensionDetails,
            configurationDetails: this.configurationDetails,
        }
        try {
            this.gqlAPIClient
                .logEvent({
                    event: eventName,
                    userCookieID: anonymousUserID,
                    source: 'IDEEXTENSION',
                    url: '',
                    argument: JSON.stringify(argument),
                    publicArgument: JSON.stringify(publicArgument),
                })
                .then(() => {})
                .catch(() => {})
        } catch (error) {
            console.log(error)
        }
    }
}
