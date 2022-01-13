import * as vscode from 'vscode'

import { endpointHostnameSetting } from './endpointSetting'
import { readConfiguration } from './readConfiguration'

const invalidAccessTokens = new Set<string>()

export function accessTokenSetting(): string | undefined {
    const fromSettings = readConfiguration().get<string>('accessToken')
    if (fromSettings) {
        return fromSettings
    }

    return undefined
}

// Ensure that only one access token error message is shown at a time.
let showingAccessTokenErrorMessage = false

export async function handleAccessTokenError(badToken: string): Promise<void> {
    invalidAccessTokens.add(badToken)

    const currentValue = readConfiguration().get<string>('accessToken')

    if (currentValue === badToken && !showingAccessTokenErrorMessage) {
        // TODO don't worry about deleting access token. Instead, prompt user to follow
        // onboarding flow in the sidebar.
        // To do this we need to maintain some type of `invalidAccessToken` state in the extension
        // and communicate or expose that to the sidebar. On access token error, show error message
        // and trigger onboarding flow. NOTE that this should work for auth sidebar validation as well
        // (e.g. user inputs bad token, we show error message and keep sidebar in "auth onboarding" state.)

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
