import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { EditorReplacements, EditorSettings } from './editor-settings'

interface LegacySettings {
    basePath?: string
    linuxBasePath?: string
    macBasePath?: string
    windowsBasePath?: string
    replacements?: EditorReplacements
    editorId?: string
    customUrlPattern?: string
    vscodeUseInsiders?: boolean
    vscodeUseSSH?: boolean
    vscodeRemoteHost?: string
    jetbrainsForceApi?: 'builtInServer'
}

/**
 * Leaves the original Settings object unchanged.
 * Returns a shallow copy with the old settings removed and the new ones added.
 */
export function migrateLegacySettings(settings: Settings): Settings {
    if (settings.openInEditor !== undefined) {
        return settings
    }

    // Migrate settings
    const legacySettings = readLegacySettingsForAllExtensions(settings)
    const legacyEditorSettingsInNewFormat = convertLegacySettingsToNewFormat(legacySettings)
    const newSettings: Settings = {
        ...settings,
        openInEditor: legacyEditorSettingsInNewFormat,
    }

    // Delete migrated legacy settings
    for (const key of Object.keys(newSettings)) {
        if (
            key.startsWith('openineditor.') ||
            key.startsWith('vscode.open.') ||
            key.startsWith('openInIntellij.') ||
            key.startsWith('openInWebstorm.') ||
            key.startsWith('openInAtom.')
        ) {
            delete newSettings[key]
        }
    }

    return newSettings
}

function convertLegacySettingsToNewFormat(legacySettings: LegacySettings): EditorSettings {
    return {
        ...(legacySettings.editorId ? { editorIds: [legacySettings.editorId] } : {}),
        ...(legacySettings.basePath ? { 'projectPaths.default': legacySettings.basePath } : {}),
        ...(legacySettings.linuxBasePath ? { 'projectPaths.linux': legacySettings.linuxBasePath } : {}),
        ...(legacySettings.macBasePath ? { 'projectPaths.mac': legacySettings.macBasePath } : {}),
        ...(legacySettings.windowsBasePath ? { 'projectPaths.windows': legacySettings.windowsBasePath } : {}),
        ...(legacySettings.replacements ? { replacements: legacySettings.replacements } : {}),
        ...(legacySettings.jetbrainsForceApi ? { 'jetbrains.forceApi': legacySettings.jetbrainsForceApi } : {}),
        ...(legacySettings.vscodeUseInsiders ? { 'vscode.useInsiders': legacySettings.vscodeUseInsiders } : {}),
        ...(legacySettings.vscodeUseSSH ? { 'vscode.useSSH': legacySettings.vscodeUseSSH } : {}),
        ...(legacySettings.vscodeRemoteHost ? { 'vscode.remoteHostForSSH': legacySettings.vscodeRemoteHost } : {}),
        ...(legacySettings.customUrlPattern ? { 'custom.urlPattern': legacySettings.customUrlPattern } : {}),
    }
}

function readLegacySettingsForAllExtensions(settings: Settings): LegacySettings {
    return {
        ...readLegacySettingsForOneExtension(settings, 'openInAtom'),
        ...readLegacySettingsForOneExtension(settings, 'openInWebstorm'),
        ...{
            ...readLegacySettingsForOneExtension(settings, 'openInIntellij'),
            jetbrainsForceApi: settings['openInIntellij.useBuiltin'] ? 'builtInServer' : undefined,
        },
        ...{
            ...readLegacySettingsForOneExtension(settings, 'vscode.open'),
            vscodeUseInsiders: settings['vscode.open.useMode'] === 'insiders',
            vscodeUseSSH: settings['vscode.open.useMode'] === 'ssh',
            vscodeRemoteHost: settings['vscode.open.remoteHost'] as string | undefined,
        },
        ...{
            ...readLegacySettingsForOneExtension(settings, 'openineditor'),
            editorId: settings['openineditor.editor'] as string | undefined,
            customUrlPattern: settings['openineditor.customUrlPattern'] as string | undefined,
        },
    }
}

function readLegacySettingsForOneExtension(settings: Settings, prefix: string): LegacySettings {
    return {
        ...(settings[prefix + '.basePath'] ? { basePath: settings[prefix + '.basePath'] as string } : null),
        ...(settings[prefix + '.osPaths.linux']
            ? { linuxBasePath: settings[prefix + '.osPaths.linux'] as string }
            : null),
        ...(settings[prefix + '.osPaths.mac'] ? { macBasePath: settings[prefix + '.osPaths.mac'] as string } : null),
        ...(settings[prefix + '.osPaths.windows']
            ? { windowsBasePath: settings[prefix + '.osPaths.windows'] as string }
            : null),
        ...(settings[prefix + '.replacements']
            ? { replacements: settings[prefix + '.replacements'] as EditorReplacements }
            : null),
    }
}
