import * as vscode from 'vscode'

import { readConfiguration } from './readConfiguration'
import { endpointHostnameSetting, endpointProtocolSetting } from './endpointSetting'
import { getInstanceVersionNumber } from '../backend/instanceVersion'

export function accessTokenSetting(): string | undefined {
    return readConfiguration().get<string>('accessToken')
}

// Ensure that only one access token error message is shown at a time.
let showingAccessTokenErrorMessage = false

export async function handleAccessTokenError(badToken?: string, endpointURL?: string): Promise<void> {
    const currentValue = readConfiguration().get<string>('accessToken')

    if (currentValue === badToken && !showingAccessTokenErrorMessage) {
        showingAccessTokenErrorMessage = true

        const message = !badToken
            ? `A valid access token is required to connect to ${endpointURL}`
            : `Connection to ${endpointURL} failed because the token is invalid. Please reload VS Code if your Sourcegraph instance URL has changed.`

        const instanceVersion = await getInstanceVersionNumber()
        const supportsTokenCallback = instanceVersion && instanceVersion < '3410'
        const action = await vscode.window.showErrorMessage(message, 'Get Token', 'Open Settings')

        if (action === 'Open Settings') {
            vscode.commands.executeCommand('workbench.action.openSettings', 'sourcegraph.accessToken')
        } else if (action === 'Get Token') {
            const path = supportsTokenCallback ? '/user/settings/tokens/new/callback' : '/user/settings/'
            const query = supportsTokenCallback ? 'requestFrom=VSCEAUTH' : ''

            vscode.env.openExternal(
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

export async function updateAccessTokenSetting(newToken: string): Promise<boolean> {
    // TODO: STORE TOKEN IN KEYCHAIN AND REMOVE FROM USER CONFIG
    try {
        await readConfiguration().update('accessToken', newToken, vscode.ConfigurationTarget.Global)
        return true
    } catch {
        return false
    }
}
