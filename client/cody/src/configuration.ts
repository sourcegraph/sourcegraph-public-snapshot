import * as vscode from 'vscode'

import type { ConfigurationUseContext, Configuration } from '@sourcegraph/cody-shared/src/configuration'

export function getConfiguration(config: vscode.WorkspaceConfiguration): Configuration {
    return {
        enabled: config.get('cody.enabled', true),
        serverEndpoint: config.get('cody.serverEndpoint', ''),
        codebase: config.get('cody.codebase'),
        debug: config.get('cody.debug', false),
        useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
        experimentalSuggest: config.get('cody.experimental.suggestions', false),
        openaiKey: config.get('cody.experimental.keys.openai', null),
    }
}

const codyConfiguration = vscode.workspace.getConfiguration('cody')
const globalConfigTarget = vscode.ConfigurationTarget.Global

// Update user configurations in VS Code for Cody
export async function updateConfiguration(configKey: string, configValue: string): Promise<void> {
    // Removing globalConfigTarget will only update configs for the workspace setting only
    await codyConfiguration.update(configKey, configValue, globalConfigTarget)
}
