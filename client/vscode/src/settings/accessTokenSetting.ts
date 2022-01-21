import * as vscode from 'vscode'

import { endpointHostnameSetting } from './endpointSetting'
import { readConfiguration } from './readConfiguration'

const invalidAccessTokens = new Set<string>()

export function accessTokenSetting(): string | undefined {
    return readConfiguration().get<string>('accessToken')
}

// Ensure that only one access token error message is shown at a time.
let showingAccessTokenErrorMessage = false

export async function handleAccessTokenError(badToken: string): Promise<void> {
    invalidAccessTokens.add(badToken)

    const currentValue = readConfiguration().get<string>('accessToken')

    if (currentValue === badToken && !showingAccessTokenErrorMessage) {
        showingAccessTokenErrorMessage = true
        await vscode.window.showErrorMessage('Invalid Sourcegraph Access Token', {
            modal: true,
            detail: `The server at ${endpointHostnameSetting()} is unable to use the access token ${badToken}.`,
        })
        showingAccessTokenErrorMessage = false
    }
}

export async function updateAccessTokenSetting(newToken: string): Promise<boolean> {
    try {
        await readConfiguration().update('accessToken', newToken, vscode.ConfigurationTarget.Global)
        return true
    } catch {
        return false
    }
}
