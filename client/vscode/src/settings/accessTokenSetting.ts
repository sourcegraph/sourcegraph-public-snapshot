import * as vscode from 'vscode'

import { isOlderThan, observeInstanceVersionNumber } from '../backend/instanceVersion'
import { extensionContext } from '../extension'
import { secretTokenKey } from '../webview/platform/AuthProvider'

import { endpointHostnameSetting, endpointProtocolSetting } from './endpointSetting'

// IMPORTANT: Call this function only once when extention is first activated
export async function processOldToken(secretStorage: vscode.SecretStorage): Promise<void> {
    // Process the token that used to live in user configuration
    // Move it to secrets and then remove it from user configuration
    const storageToken = await secretStorage.get(secretTokenKey)
    const oldToken = vscode.workspace.getConfiguration().get<string>('sourcegraph.accessToken') || ''
    if (!storageToken && oldToken.length > 8) {
        await secretStorage.store(secretTokenKey, oldToken)
        await vscode.workspace
            .getConfiguration()
            .update('sourcegraph.accessToken', undefined, vscode.ConfigurationTarget.Global)
        await vscode.workspace
            .getConfiguration()
            .update('sourcegraph.accessToken', undefined, vscode.ConfigurationTarget.Workspace)
    }
    return
}

export async function getAccessToken(): Promise<string | undefined> {
    const token = await extensionContext?.secrets.get(secretTokenKey)
    return token
}

// Ensure that only one access token error message is shown at a time.
let showingAccessTokenErrorMessage = false

export async function handleAccessTokenError(badToken: string, endpointURL: string): Promise<void> {
    if (badToken !== undefined && !showingAccessTokenErrorMessage) {
        showingAccessTokenErrorMessage = true

        const message = !badToken
            ? `A valid access token is required to connect to ${endpointURL}`
            : `Connection to ${endpointURL} failed. Please check your access token and network connection.`

        const version = await observeInstanceVersionNumber(badToken, endpointURL).toPromise()
        const supportsTokenCallback = version && isOlderThan(version, { major: 3, minor: 41 })
        const action = await vscode.window.showErrorMessage(message, 'Get Token', 'Reload Window')

        if (action === 'Reload Window') {
            await vscode.commands.executeCommand('workbench.action.reloadWindow')
        } else if (action === 'Get Token') {
            const path = supportsTokenCallback ? '/user/settings/tokens/new/callback' : '/user/settings/'
            const query = supportsTokenCallback ? 'requestFrom=VSCEAUTH' : ''

            await vscode.env.openExternal(
                vscode.Uri.from({
                    scheme: endpointProtocolSetting().slice(0, -1),
                    authority: endpointHostnameSetting(),
                    path,
                    query,
                })
            )
        }
        showingAccessTokenErrorMessage = false
    }
}
