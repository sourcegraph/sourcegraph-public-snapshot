import * as path from 'path'

import type { UIPositionSpec } from '@sourcegraph/shared/src/util/url'

import { ExternalServiceKind } from '../graphql-operations'

import type { EditorReplacements, EditorSettings } from './editor-settings'
import { type Editor, getEditor, supportedEditors } from './editors'

// Just lowercasing these for now, it's a bit of a gamble because it's only a coincidence that ExternalServiceKind
// and ExternalServiceType only differs in casing. But it works for now.
// Discussion about this here: https://github.com/sourcegraph/sourcegraph/pull/41914#discussion_r977418118
const serviceTypesWithOwnerInUrl = new Set<string>([
    ExternalServiceKind.GITHUB.toLowerCase(),
    ExternalServiceKind.GITLAB.toLowerCase(),
    ExternalServiceKind.BITBUCKETCLOUD.toLowerCase(),
])

export function buildRepoBaseNameAndPath(
    repoName: string,
    externalServiceType: string | undefined,
    filePath: string | undefined
): string {
    const pathPieces = repoName.split('/')
    const shouldCutOffTwoPieces =
        pathPieces.length > 2 && serviceTypesWithOwnerInUrl.has((externalServiceType || '').toLowerCase())
    const bareRepoNamePieces = pathPieces.slice(shouldCutOffTwoPieces ? 2 : 1)
    return path.join(...bareRepoNamePieces, ...(filePath ? [filePath] : []))
}

export function buildEditorUrl(
    repoBaseNameAndPath: string,
    position: UIPositionSpec['position'] | undefined,
    editorSettings: EditorSettings | undefined,
    sourcegraphBaseUrl: string,
    editorIndex = 0
): URL {
    const editorSettingsErrorMessage = getEditorSettingsErrorMessage(editorSettings, sourcegraphBaseUrl)
    if (editorSettingsErrorMessage) {
        throw new TypeError(editorSettingsErrorMessage)
    }

    const projectPath = getProjectPath(editorSettings || {}) as string
    const editor = getEditor(editorSettings?.editorIds?.[editorIndex] ?? 'vscode') as Editor
    const urlPattern = getUrlPattern(editor, editorSettings || {})
    // If VS Code && (Windows || UNC flag is on), add an extra slash in the beginning
    const pathPrefix =
        editor.id === 'vscode' && (isWindowsPath(projectPath) || editorSettings?.['vscode.isProjectPathUNCPath'])
            ? '/'
            : ''

    const absolutePath = path.join(projectPath, repoBaseNameAndPath)
    const { line, column } = position ? { line: position.line, column: position.character } : { line: 1, column: 1 }
    const url = urlPattern
        .replace('%file', pathPrefix + absolutePath)
        .replace('%line', `${line}`)
        .replace('%col', `${column}`)
    return new URL(doReplacements(url, editorSettings?.replacements))
}

export function getEditorSettingsErrorMessage(
    editorSettings: EditorSettings | undefined,
    sourcegraphBaseUrl: string
): string | undefined {
    const learnMoreURL = 'https://sourcegraph.com/docs/integration/open_in_editor'

    if (!editorSettings) {
        return `Add \`openInEditor\` to your user settings to open files in the editor. [Learn more](${learnMoreURL})`
    }

    const projectPath = getProjectPath(editorSettings)

    if (typeof projectPath !== 'string') {
        return `Add \`projectPaths.default\` or some OS-specific path to your user settings to open files in the editor. [Learn more](${learnMoreURL})`
    }

    // Skip this check on Windows because path.isAbsolute only checks Linux and macOS compatible paths reliably
    if (!isProjectPathValid(projectPath)) {
        return `\`projectPaths.default\` (or your current OS-specific setting) \`${projectPath}\` is not an absolute path. Please correct the error in your [user settings](${
            new URL('/user/settings', sourcegraphBaseUrl).href
        }).`
    }

    if (!editorSettings.editorIds?.length) {
        return `Add \`editorIds\` to your user settings to open files. [Learn more](${learnMoreURL})`
    }
    const validEditorCount = editorSettings.editorIds.map(id => getEditor(id)).filter(editor => editor).length

    if (validEditorCount !== editorSettings.editorIds.length) {
        return (
            `Setting \`editorIds\` must be set to a valid array of values in your [user settings](${
                new URL('/user/settings', sourcegraphBaseUrl).href
            }) to open files. Supported editors: ` + supportedEditors.map(editor => editor.id).join(', ')
        )
    }
    if (editorSettings.editorIds?.includes('custom') && typeof editorSettings['custom.urlPattern'] !== 'string') {
        return `Add \`custom.urlPattern\` to your user settings for custom editor to open files. [Learn more](${learnMoreURL})`
    }

    if (editorSettings['vscode.useSSH'] && !editorSettings['vscode.remoteHostForSSH']) {
        return '`vscode.useSSH` is set to "true" but `vscode.remoteHostForSSH` is not set.'
    }

    return undefined
}

export function isProjectPathValid(projectPath: string | undefined): boolean {
    return !!projectPath && (isWindowsPath(projectPath) || path.isAbsolute(projectPath))
}

export function getProjectPath(editorSettings: EditorSettings): string | undefined {
    if (navigator.userAgent.includes('Win') && editorSettings['projectPaths.windows']) {
        return editorSettings['projectPaths.windows']
    }
    if (navigator.userAgent.includes('Mac') && editorSettings['projectPaths.mac']) {
        return editorSettings['projectPaths.mac']
    }
    if (navigator.userAgent.includes('Linux') && editorSettings['projectPaths.linux']) {
        return editorSettings['projectPaths.linux']
    }
    return editorSettings['projectPaths.default']
}

function isWindowsPath(path: string): boolean {
    return /^[A-Za-z]:\\/.test(path)
}

export function getUrlPattern(editor: Editor, editorSettings: EditorSettings): string {
    if (editor.urlPattern) {
        return editor.urlPattern
    }
    if (editor.id === 'vscode') {
        const protocolHandler = editorSettings['vscode.useInsiders'] ? 'vscode-insiders' : 'vscode'
        if (editorSettings['vscode.useSSH']) {
            if (!editorSettings['vscode.remoteHostForSSH']) {
                throw new TypeError(
                    '`openineditor.vscode.mode` is set to "ssh" but `openineditor.vscode.remoteHostForSSH` is not set.'
                )
            }
            return `${protocolHandler}://vscode-remote/ssh-remote+${editorSettings['vscode.remoteHostForSSH']}%file:%line:%col`
        }
        return `${protocolHandler}://file%file:%line:%col`
    }
    if (editor.isJetBrainsProduct) {
        if (editorSettings['jetbrains.forceApi'] === 'builtInServer') {
            // Open files with IntelliJ's built-in REST API (port 63342) if useBuiltin is enabled instead of the idea:// protocol handler
            // ref: https://www.jetbrains.com/help/idea/php-built-in-web-server.html#configuring-built-in-web-server
            return 'http://localhost:63342/api/file%file:%line:%col'
        }
        return `${editor.id}://open?file=%file&line=%line&column=%col`
    }
    if (editor.id === 'custom') {
        return editorSettings['custom.urlPattern'] ?? ''
    }
    throw new TypeError(`No url pattern found for editor ${editor.id}`)
}

function doReplacements(urlWithoutReplacements: string, replacements: EditorReplacements | undefined): string {
    let url = urlWithoutReplacements
    if (replacements) {
        for (const [search, replacement] of Object.entries(replacements)) {
            url = url.replace(new RegExp(search), replacement)
        }
    }
    return url
}
