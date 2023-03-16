import * as vscode from 'vscode'

export const CODY_ENDPOINT = 'cody.sgdev.org'

export const CODY_ACCESS_TOKEN_SECRET = 'cody.access-token'

export type ConfigurationUseContext = 'embeddings' | 'keyword' | 'none' | 'blended'

export interface Configuration {
    enable: boolean
    serverEndpoint: string
    embeddingsEndpoint: string
    codebase?: string
    debug: boolean
    useContext: ConfigurationUseContext
    experimentalSuggest: boolean
}

export function getConfiguration(config: vscode.WorkspaceConfiguration): Configuration {
    return {
        enable: config.get('sourcegraph.cody.enable', true),
        // FIXME: Remove these lint suppressions when we can supply a default endpoint
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        serverEndpoint: config.get('cody.serverEndpoint')!,
        // FIXME: Remove these lint suppressions when we can supply a default endpoint
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        embeddingsEndpoint: config.get('cody.embeddingsEndpoint')!,
        codebase: config.get('cody.codebase'),
        debug: config.get('cody.debug', false),
        useContext: config.get<ConfigurationUseContext>('cody.useContext') || 'embeddings',
        experimentalSuggest: config.get('cody.experimental.suggest', false),
    }
}

const codyConfiguration = vscode.workspace.getConfiguration('cody')
const globalConfigTarget = vscode.ConfigurationTarget.Global

// Update user configurations in VS Code for Cody
export async function updateConfiguration(configKey: string, configValue: string): Promise<void> {
    // Removing globalConfigTarget will only update configs for the workspace setting only
    await codyConfiguration.update(configKey, configValue, globalConfigTarget)
}
