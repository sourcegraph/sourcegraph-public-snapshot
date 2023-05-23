import * as vscode from 'vscode'

import type {
    ConfigurationUseContext,
    Configuration,
    ConfigurationWithAccessToken,
} from '@sourcegraph/cody-shared/src/configuration'

import { SecretStorage, getAccessToken } from './services/SecretStorageProvider'

/**
 * All configuration values, with some sanitization performed.
 */
export function getConfiguration(config: Pick<vscode.WorkspaceConfiguration, 'get'>): Configuration {
    let debugRegex: RegExp | null = null
    try {
        const debugPattern: string | null = config.get<string | null>('cody.debug.filter', null)
        if (debugPattern) {
            if (debugPattern === '*') {
                debugRegex = new RegExp('.*')
            } else {
                debugRegex = new RegExp(debugPattern)
            }
        }
    } catch (error) {
        void vscode.window.showErrorMessage("Error parsing cody.debug.filter regex - using default '*'", error)
        debugRegex = new RegExp('.*')
    }

    return {
        serverEndpoint: sanitizeServerEndpoint(config.get('cody.serverEndpoint', '')),
        codebase: sanitizeCodebase(config.get('cody.codebase')),
        debugEnable: config.get<boolean>('cody.debug.enable', false),
        debugVerbose: config.get<boolean>('cody.debug.verbose', false),
        debugFilter: debugRegex,
        useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
        experimentalSuggest: config.get('cody.experimental.suggestions', false),
        experimentalChatPredictions: config.get('cody.experimental.chatPredictions', false),
        experimentalInline: config.get('cody.experimental.inline', false),
        experimentalGuardrails: config.get('cody.experimental.guardrails', false),
        customHeaders: config.get<object>('cody.customHeaders', {}) as Record<string, string>,
    }
}

function sanitizeCodebase(codebase: string | undefined): string {
    if (!codebase) {
        return ''
    }
    const protocolRegexp = /^(https?):\/\//
    const trailingSlashRegexp = /\/$/
    return codebase.replace(protocolRegexp, '').trim().replace(trailingSlashRegexp, '')
}

function sanitizeServerEndpoint(serverEndpoint: string): string {
    const trailingSlashRegexp = /\/$/
    return serverEndpoint.trim().replace(trailingSlashRegexp, '')
}

const codyConfiguration = vscode.workspace.getConfiguration('cody')

// Update user configurations in VS Code for Cody
export async function updateConfiguration(configKey: string, configValue: string): Promise<void> {
    await codyConfiguration.update(configKey, configValue, vscode.ConfigurationTarget.Global)
}

export const getFullConfig = async (secretStorage: SecretStorage): Promise<ConfigurationWithAccessToken> => {
    const config = getConfiguration(vscode.workspace.getConfiguration())
    const accessToken = (await getAccessToken(secretStorage)) || null
    return { ...config, accessToken }
}
