import * as vscode from 'vscode'

import { isOlderThan, observeInstanceVersionNumber } from '../backend/instanceVersion'
import { scretTokenKey } from '../webview/platform/AuthProvider'

import { endpointHostnameSetting, endpointProtocolSetting } from './endpointSetting'
import { readConfiguration } from './readConfiguration'

export async function accessTokenSetting(secretStorage: vscode.SecretStorage): Promise<string> {
    const currentToken = await secretStorage.get(scretTokenKey)
    return currentToken || ''
}

export async function removeAccessTokenSetting(): Promise<void> {
    await readConfiguration().update('accessToken', undefined, vscode.ConfigurationTarget.Global)
    await readConfiguration().update('accessToken', undefined, vscode.ConfigurationTarget.Workspace)
    return
}

// Ensure that only one access token error message is shown at a time.
let showingAccessTokenErrorMessage = false

export async function handleAccessTokenError(badToken: string, endpointURL: string): Promise<void> {
    if (badToken !== undefined && !showingAccessTokenErrorMessage) {
        showingAccessTokenErrorMessage = true

        const message = !badToken
            ? `A valid access token is required to connect to ${endpointURL}`
            : `Connection to ${endpointURL} failed because the token is invalid. Please reload VS Code if your Sourcegraph instance URL has changed.`

        const version = await observeInstanceVersionNumber(badToken, endpointURL).toPromise()
        const supportsTokenCallback = version && isOlderThan(version, { major: 3, minor: 41 })
        const action = await vscode.window.showErrorMessage(message, 'Get Token', 'Update URL in Setting')

        if (action === 'Open Settings') {
            await vscode.commands.executeCommand('workbench.action.openSettings', 'sourcegraph.url')
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
