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
    const oldEditorSettings = {
        ...readAtomSettings(settings),
        ...readWebStormSettings(settings),
        ...readIntelliJSettings(settings),
        ...readVSCodeSettings(settings),
        ...readOpenInEditorExtensionSettings(settings),
    }

    const mergedEditorSettings: EditorSettings = {
        ...(oldEditorSettings.editorId ? { editorId: oldEditorSettings.editorId } : {}),
        projectPaths: {
            ...(oldEditorSettings.basePath ? { default: oldEditorSettings.basePath } : {}),
            ...(oldEditorSettings.linuxBasePath ? { linux: oldEditorSettings.linuxBasePath } : {}),
            ...(oldEditorSettings.macBasePath ? { mac: oldEditorSettings.macBasePath } : {}),
            ...(oldEditorSettings.windowsBasePath ? { windows: oldEditorSettings.windowsBasePath } : {}),
        },
        ...(oldEditorSettings.replacements ? { replacements: oldEditorSettings.replacements } : {}),
        ...(oldEditorSettings.jetbrainsForceApi
            ? { jetbrains: { forceApi: oldEditorSettings.jetbrainsForceApi } }
            : {}),
        vscode: {
            ...(oldEditorSettings.vscodeUseInsiders ? { useInsiders: oldEditorSettings.vscodeUseInsiders } : {}),
            ...(oldEditorSettings.vscodeUseSSH ? { useSSH: oldEditorSettings.vscodeUseSSH } : {}),
            ...(oldEditorSettings.vscodeRemoteHost ? { remoteHostForSSH: oldEditorSettings.vscodeRemoteHost } : {}),
        },
        ...(oldEditorSettings.customUrlPattern ? { custom: { urlPattern: oldEditorSettings.customUrlPattern } } : {}),
    }
    if (!Object.keys({ ...mergedEditorSettings.projectPaths }).length) {
        delete mergedEditorSettings.projectPaths
    }
    if (!Object.keys({ ...mergedEditorSettings.vscode }).length) {
        delete mergedEditorSettings.projectPaths
    }

    return mergedEditorSettings
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
    const osPaths = (settings[prefix + '.osPaths'] as Record<string, string>) || {}
    return {
        basePath: settings[prefix + '.basePath'] as string | undefined,
        linuxBasePath: osPaths.linux as string | undefined,
        macBasePath: osPaths.mac as string | undefined,
        windowsBasePath: osPaths.windows as string | undefined,
        replacements: settings[prefix + '.replacements'] as EditorReplacements | undefined,
    }
}
