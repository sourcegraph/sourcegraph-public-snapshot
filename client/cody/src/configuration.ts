import * as vscode from 'vscode'

import type {
    ConfigurationUseContext,
    Configuration,
    ConfigurationWithAccessToken,
} from '@sourcegraph/cody-shared/src/configuration'

import { SecretStorage, getAccessToken, getServerEndpoint } from './services/SecretStorageProvider'

/**
 * All configuration values, with some sanitization performed.
 */
export function getConfiguration(
    config: Pick<vscode.WorkspaceConfiguration, 'get'>
): Omit<Configuration, 'serverEndpoint'> {
    return {
        codebase: sanitizeCodebase(config.get('cody.codebase')),
        debug: config.get('cody.debug', false),
        useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
        experimentalSuggest: config.get('cody.experimental.suggestions', false),
        experimentalChatPredictions: config.get('cody.experimental.chatPredictions', false),
        experimentalInline: config.get('cody.experimental.inline', false),
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

const codyConfiguration = vscode.workspace.getConfiguration('cody')

// Update user configurations in VS Code for Cody
export async function updateConfiguration(configKey: string, configValue: string): Promise<void> {
    await codyConfiguration.update(configKey, configValue, vscode.ConfigurationTarget.Global)
}

export const getFullConfig = async (secretStorage: SecretStorage): Promise<ConfigurationWithAccessToken> => {
    const config = getConfiguration(vscode.workspace.getConfiguration())
    const accessToken = (await getAccessToken(secretStorage)) || null
    const serverEndpoint = await getServerEndpoint(secretStorage)
    return { ...config, accessToken, serverEndpoint }
}
