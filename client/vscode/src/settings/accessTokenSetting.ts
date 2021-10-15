import { once } from 'lodash'
import open from 'open'
import * as vscode from 'vscode'

import { log } from '../log'

import { endpointHostnameSetting, endpointSetting } from './endpointSetting'
import { readConfiguration } from './readConfiguration'

const invalidAccessTokens = new Set<string>()

export function accessTokenSetting(): string | undefined {
    const fromSettings = readConfiguration().get<string>('accessToken')
    if (fromSettings) {
        return fromSettings
    }

    const environmentVariable = process.env.SRC_ACCESS_TOKEN
    if (environmentVariable && !invalidAccessTokens.has(environmentVariable)) {
        return environmentVariable
    }

    return undefined
}

export async function handleAccessTokenError(tokenValueToDelete: string): Promise<void> {
    invalidAccessTokens.add(tokenValueToDelete)

    const currentValue = readConfiguration().get<string>('accessToken')
    if (currentValue === tokenValueToDelete) {
        await readConfiguration().update('accessToken', undefined, vscode.ConfigurationTarget.Global)

        await promptUserForAccessTokenSetting(tokenValueToDelete)
    } else {
        log.appendLine(
            `can't delete access token '${tokenValueToDelete}' because it doesn't match ` +
                `existing configuration value '${currentValue || 'undefined'}'`
        )
    }
}

// Ensure only one prompt at a time (likely multiple request failures at once, only need to prompt for token once,
// extension will not work until VS Code is reloaded).
const promptUserForAccessTokenSetting = once(
    async (badToken: string): Promise<string | undefined> => {
        try {
            const title = 'Invalid Sourcegraph Access Token'
            const detail = `The server at ${endpointHostnameSetting()} is unable to use the access token ${badToken}.`

            const openBrowserMessage = 'Open browser to create an access token'
            const logout = 'Continue without an access token'
            const userChoice = await vscode.window.showErrorMessage(
                title,
                { modal: true, detail },
                openBrowserMessage,
                logout
            )

            if (userChoice === openBrowserMessage) {
                await open(`${endpointSetting()}/user/settings/tokens`)
                const newToken = await vscode.window.showInputBox({
                    title: 'Paste your Sourcegraph access token here',
                    ignoreFocusOut: true,
                })
                if (newToken) {
                    await readConfiguration().update('accessToken', newToken, vscode.ConfigurationTarget.Global)

                    // TODO flesh out onboarding flow.
                    await vscode.window.showInformationMessage(
                        'Updated Sourcegraph access token. Reload VS Code for this change to take effect.'
                    )
                    log.appendLine(`new access token from user: ${newToken || 'undefined'}`)

                    return newToken
                }
            }
            return undefined
        } catch (error) {
            log.error('promptUserForAccessTokenSetting', error)
            return undefined
        }
    }
)
