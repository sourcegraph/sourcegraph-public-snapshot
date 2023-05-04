import * as vscode from 'vscode'

import type {
    ConfigurationUseContext,
    Configuration,
    ConfigurationWithAccessToken,
} from '@sourcegraph/cody-shared/src/configuration'

import { SecretStorage, getAccessToken } from './secret-storage'

/**
 * All configuration values, with some sanitization performed.
 */
export function getConfiguration(config: Pick<vscode.WorkspaceConfiguration, 'get'>): Configuration {
    return {
        enabled: config.get('cody.enabled', true),
        serverEndpoint: sanitizeServerEndpoint(config.get('cody.serverEndpoint', '')),
        codebase: sanitizeCodebase(config.get('cody.codebase')),
        debug: config.get('cody.debug', false),
        useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
        experimentalSuggest: config.get('cody.experimental.suggestions', false),
        experimentalChatPredictions: config.get('cody.experimental.chatPredictions', false),
        anthropicKey: config.get('cody.experimental.keys.anthropic', null),
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
    return { ...config, accessToken: await getAccessToken(secretStorage) }
}
