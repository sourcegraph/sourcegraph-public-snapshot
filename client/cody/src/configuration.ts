import * as vscode from 'vscode'

import type {
    ConfigurationUseContext,
    Configuration,
    ConfigurationWithAccessToken,
} from '@sourcegraph/cody-shared/src/configuration'

import { CONFIG_KEY, ConfigKeys } from './configuration-keys'
import { SecretStorage, getAccessToken } from './services/SecretStorageProvider'

interface ConfigGetter {
    get<T>(section: typeof CONFIG_KEY[ConfigKeys], defaultValue?: T): T
}

/**
 * All configuration values, with some sanitization performed.
 */
export function getConfiguration(config: ConfigGetter): Configuration {
    const isTesting = process.env.CODY_TESTING === 'true'

    let debugRegex: RegExp | null = null
    try {
        const debugPattern: string | null = config.get<string | null>(CONFIG_KEY.debugFilter, null)
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

    let completionsAdvancedProvider = config.get<'anthropic' | 'unstable-codegen'>(
        CONFIG_KEY.completionsAdvancedProvider,
        'anthropic'
    )
    if (
        completionsAdvancedProvider !== 'anthropic' &&
        completionsAdvancedProvider !== 'unstable-codegen' &&
        completionsAdvancedProvider !== 'unstable-huggingface'
    ) {
        completionsAdvancedProvider = 'anthropic'
        void vscode.window.showInformationMessage(
            `Unrecognized ${CONFIG_KEY.completionsAdvancedProvider}, defaulting to 'anthropic'`
        )
    }

    return {
        serverEndpoint: sanitizeServerEndpoint(config.get(CONFIG_KEY.serverEndpoint, '')),
        codebase: sanitizeCodebase(config.get(CONFIG_KEY.codebase)),
        customHeaders: config.get<object>(CONFIG_KEY.customHeaders, {}) as Record<string, string>,
        useContext: config.get<ConfigurationUseContext>(CONFIG_KEY.useContext) || 'embeddings',
        debugEnable: config.get<boolean>(CONFIG_KEY.debugEnable, false),
        debugVerbose: config.get<boolean>(CONFIG_KEY.debugVerbose, false),
        debugFilter: debugRegex,
        autocomplete: config.get(CONFIG_KEY.autocomplete, isTesting),
        experimentalChatPredictions: config.get(CONFIG_KEY.experimentalChatPredictions, isTesting),
        experimentalInline: config.get(CONFIG_KEY.experimentalInline, isTesting),
        experimentalGuardrails: config.get(CONFIG_KEY.experimentalGuardrails, isTesting),
        experimentalNonStop: config.get(CONFIG_KEY.experimentalNonStop, isTesting),
        completionsAdvancedProvider,
        completionsAdvancedServerEndpoint: config.get<string | null>(
            CONFIG_KEY.completionsAdvancedServerEndpoint,
            null
        ),
        completionsAdvancedAccessToken: config.get<string | null>(CONFIG_KEY.completionsAdvancedAccessToken, null),
        completionsAdvancedCache: config.get(CONFIG_KEY.completionsAdvancedCache, true),
        completionsAdvancedEmbeddings: config.get(CONFIG_KEY.completionsAdvancedEmbeddings, true),
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

// We run this callback on extension startup
export async function migrateConfiguration(): Promise<void> {
    await migrateDeprecatedConfigOption(CONFIG_KEY.experimentalSuggestions, CONFIG_KEY.autocomplete)
    // await migrateDeprecatedConfigOption(CONFIG_KEY.completionsAdvancedProvider, CONFIG_KEY.debugFilter)
}

async function migrateDeprecatedConfigOption(
    oldKey: typeof CONFIG_KEY[ConfigKeys],
    newKey: typeof CONFIG_KEY[ConfigKeys]
): Promise<void> {
    const config = vscode.workspace.getConfiguration()
    const value = config.get(oldKey)
    const inspect = config.inspect(oldKey)

    if (value === undefined || inspect === undefined) {
        return
    }

    const scope =
        inspect.workspaceFolderValue !== undefined
            ? vscode.ConfigurationTarget.WorkspaceFolder
            : inspect?.workspaceValue !== undefined
            ? vscode.ConfigurationTarget.Workspace
            : vscode.ConfigurationTarget.Global

    console.log('found the setting in scope', scope, value)

    await config.update(newKey, value, scope)
    await config.update(oldKey, undefined, scope)

    // // Set the new setting value for the current scope
    // workspace.getConfiguration().update(newSettingName, oldSettingValue, getScope())

    // // Remove the old setting for all possible scopes
    // workspace.getConfiguration().update(oldSettingName, undefined)
    // workspace.getConfiguration(undefined).update(oldSettingName, undefined)
}
