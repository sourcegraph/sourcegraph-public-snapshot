import * as vscode from 'vscode'

export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none' | 'blended'

export interface Configuration {
    enabled: boolean
    serverEndpoint: string
    codebase?: string
    debug: boolean
    useContext: ConfigurationUseContext
    experimentalSuggest: boolean
    openaiKey: string | null
}

export function getConfiguration(config: vscode.WorkspaceConfiguration): Configuration {
    return {
        enabled: config.get('cody.enabled', true),
        serverEndpoint: config.get('cody.serverEndpoint', ''),
        codebase: config.get('cody.codebase'),
        debug: config.get('cody.debug', false),
        useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
        experimentalSuggest: config.get('cody.experimental.suggest', false),
        openaiKey: config.get('cody.keys.openai', null),
    }
}

const codyConfiguration = vscode.workspace.getConfiguration('cody')
const globalConfigTarget = vscode.ConfigurationTarget.Global

// Update user configurations in VS Code for Cody
export async function updateConfiguration(configKey: string, configValue: string): Promise<void> {
    // Removing globalConfigTarget will only update configs for the workspace setting only
    await codyConfiguration.update(configKey, configValue, globalConfigTarget)
}
