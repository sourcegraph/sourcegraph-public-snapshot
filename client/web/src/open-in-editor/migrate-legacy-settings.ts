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

    const oldEditorSettingsInNewFormat = getOldSettingsInNewFormat(settings)

    const newSettings: Settings = {
        ...settings,
        openInEditor: oldEditorSettingsInNewFormat,
    }
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

function getOldSettingsInNewFormat(settings: Settings): EditorSettings {
    const old = {
        ...readAtomSettings(settings),
        ...readWebStormSettings(settings),
        ...readIntelliJSettings(settings),
        ...readVSCodeSettings(settings),
        ...readOpenInEditorExtensionSettings(settings),
    }

    return {
        ...(old.editorId ? { editorId: old.editorId } : {}),
        ...(old.basePath ? { 'projectPaths.default': old.basePath } : {}),
        ...(old.linuxBasePath ? { 'projectPaths.linux': old.linuxBasePath } : {}),
        ...(old.macBasePath ? { 'projectPaths.mac': old.macBasePath } : {}),
        ...(old.windowsBasePath ? { 'projectPaths.windows': old.windowsBasePath } : {}),
        ...(old.replacements ? { replacements: old.replacements } : {}),
        ...(old.jetbrainsForceApi ? { 'jetbrains.forceApi': old.jetbrainsForceApi } : {}),
        ...(old.vscodeUseInsiders ? { 'vscode.useInsiders': old.vscodeUseInsiders } : {}),
        ...(old.vscodeUseSSH ? { 'vscode.useSSH': old.vscodeUseSSH } : {}),
        ...(old.vscodeRemoteHost ? { 'vscode.remoteHostForSSH': old.vscodeRemoteHost } : {}),
        ...(old.customUrlPattern ? { 'custom.urlPattern': old.customUrlPattern } : {}),
    }
}

function readOpenInEditorExtensionSettings(settings: Settings): LegacySettings {
    return {
        ...readLegacySettings(settings, 'openineditor'),
        editorId: settings['openineditor.editor'] as string | undefined,
        customUrlPattern: settings['openineditor.customUrlPattern'] as string | undefined,
    }
}

function readVSCodeSettings(settings: Settings): LegacySettings {
    return {
        ...readLegacySettings(settings, 'vscode.open'),
        vscodeUseInsiders: settings['vscode.open.useMode'] === 'insiders',
        vscodeUseSSH: settings['vscode.open.useMode'] === 'ssh',
        vscodeRemoteHost: settings['vscode.open.remoteHost'] as string | undefined,
    }
}

function readIntelliJSettings(settings: Settings): LegacySettings {
    return {
        ...readLegacySettings(settings, 'openInIntellij'),
        jetbrainsForceApi: settings['openInIntellij.useBuiltin'] ? 'builtInServer' : undefined,
    }
}

function readWebStormSettings(settings: Settings): LegacySettings {
    return readLegacySettings(settings, 'openInWebstorm')
}

function readAtomSettings(settings: Settings): LegacySettings {
    return readLegacySettings(settings, 'openInAtom')
}

function readLegacySettings(settings: Settings, prefix: string): LegacySettings {
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
