import * as vscode from 'vscode'

import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

function getServerEndpointFromConfig(config: vscode.WorkspaceConfiguration): string {
    return config.get<string>('cody.serverEndpoint', '')
}

function getUseContextFromConfig(config: vscode.WorkspaceConfiguration): string {
    if (!config) {
        return ''
    }
    return config.get<string>('cody.useContext', '')
}

function getChatPredictionsFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    if (!config) {
        return false
    }
    return config.get<boolean>('cody.experimental.chatPredictions', false)
}

function getInlineFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    if (!config) {
        return false
    }
    return config.get<boolean>('cody.experimental.inline', false)
}

function getNonStopFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    if (!config) {
        return false
    }
    return config.get<boolean>('cody.experimental.nonStop', false)
}

function getSuggestionsFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    if (!config) {
        return false
    }
    return config.get<boolean>('cody.experimental.suggestions', false)
}

function getGuardrailsFromConfig(config: vscode.WorkspaceConfiguration): boolean {
    if (!config) {
        return false
    }
    return config.get<boolean>('cody.experimental.guardrails', false)
}

export class EventLogger {
    private serverEndpoint = getServerEndpointFromConfig(vscode.workspace.getConfiguration())
    private extensionDetails = { ide: 'VSCode', ideExtensionType: 'Cody' }
    private constructor(private gqlAPIClient: SourcegraphGraphQLAPIClient) {}

    public static create(gqlAPIClient: SourcegraphGraphQLAPIClient): EventLogger {
        return new EventLogger(gqlAPIClient)
    }

    public configurationDetails = {
        contextSelection: getUseContextFromConfig(vscode.workspace.getConfiguration()),
        chatPredictions: getChatPredictionsFromConfig(vscode.workspace.getConfiguration()),
        inline: getInlineFromConfig(vscode.workspace.getConfiguration()),
        nonStop: getNonStopFromConfig(vscode.workspace.getConfiguration()),
        suggestions: getSuggestionsFromConfig(vscode.workspace.getConfiguration()),
        guardrails: getGuardrailsFromConfig(vscode.workspace.getConfiguration()),
    }

    public onConfigurationChange(newconfig: vscode.WorkspaceConfiguration): void {
        this.configurationDetails = {
            contextSelection: getUseContextFromConfig(newconfig),
            chatPredictions: getChatPredictionsFromConfig(newconfig),
            inline: getInlineFromConfig(newconfig),
            nonStop: getNonStopFromConfig(newconfig),
            suggestions: getSuggestionsFromConfig(newconfig),
            guardrails: getGuardrailsFromConfig(newconfig),
        }
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
