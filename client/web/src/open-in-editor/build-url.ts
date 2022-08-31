import * as path from 'path'

import type { UIRangeSpec } from '@sourcegraph/shared/out/src/util/url'

import type { EditorReplacements, EditorSettings } from './editor-settings'
import { Editor, getEditor, supportedEditors } from './editors'

export function buildEditorUrl(
    repoBaseNameAndPath: string,
    range: UIRangeSpec['range'] | undefined,
    editorSettings: EditorSettings,
    sourcegraphBaseUrl: string
): URL {
    const basePath = getProjectPathWithChecks(editorSettings, sourcegraphBaseUrl)
    const editor = getEditorWithChecks(editorSettings.editorId, editorSettings.custom?.urlPattern, sourcegraphBaseUrl)
    const urlPattern = getUrlPattern(editor, editorSettings)
    // If VS Code && (Windows || UNC flag is on), add an extra slash in the beginning
    const pathPrefix =
        editor.id === 'vscode' && (/^[A-Za-z]:\\/.test(basePath) || editorSettings.vscode?.isBasePathUNCPath) ? '/' : ''

    const absolutePath = path.join(basePath, repoBaseNameAndPath)
    const { line, column } = range ? { line: range.start.line, column: range.start.character } : { line: 1, column: 1 }
    const url = urlPattern
        .replace('%file', pathPrefix + absolutePath)
        .replace('%line', `${line}`)
        .replace('%col', `${column}`)
    return new URL(doReplacements(url, editorSettings.replacements))
}

function getProjectPathWithChecks(editorSettings: EditorSettings, sourcegraphBaseUrl: string): string {
    const basePath = getProjectPath(editorSettings)

    if (typeof basePath !== 'string') {
        throw new TypeError( // TODO: Improve this error message with OS-specific things
            `Add \`projectsPaths.default\` to your user settings to open files in the editor. [Learn more](${getLearnMorePath(
                sourcegraphBaseUrl
            )})`
        )
    }
    if (!path.isAbsolute(basePath)) {
        throw new Error( // TODO: Improve this error message with OS-specific things
            `\`projectsPaths.default\` value \`${basePath}\` is not an absolute path. Please correct the error in your [user settings](${
                new URL('/user/settings', sourcegraphBaseUrl).href
            }).`
        )
    }

    return basePath
}

function getProjectPath(editorSettings: EditorSettings): string | undefined {
    if (editorSettings.projectsPaths) {
        if (navigator.userAgent.includes('Win') && editorSettings.projectsPaths.windows) {
            return editorSettings.projectsPaths.windows
        }
        if (navigator.userAgent.includes('Mac') && editorSettings.projectsPaths.mac) {
            return editorSettings.projectsPaths.mac
        }
        if (navigator.userAgent.includes('Linux') && editorSettings.projectsPaths.linux) {
            return editorSettings.projectsPaths.linux
        }
    }
    return editorSettings.projectsPaths?.default
}

function getEditorWithChecks(
    editorId: string | undefined,
    urlPattern: string | undefined,
    sourcegraphBaseUrl: string
): Editor {
    const learnMorePath = getLearnMorePath(sourcegraphBaseUrl)

    if (typeof editorId !== 'string') {
        throw new TypeError(
            `Add \`openineditor.editorId\` to your user settings to open files. [Learn more](${learnMorePath})`
        )
    }
    const editor = getEditor(editorId)

    if (!editor) {
        throw new TypeError(
            `Setting \`openineditor.editorId\` must be set to a valid value in your [user settings](${
                new URL('/user/settings', sourcegraphBaseUrl).href
            }) to open files. Supported editors: ` + supportedEditors.map(editor => editor.id).join(', ')
        )
    }
    if (editorId === 'custom' && typeof urlPattern !== 'string') {
        throw new TypeError(
            `Add \`openineditor.customUrlPattern\` to your user settings for custom editor to open files. [Learn more](${learnMorePath})`
        )
    }

    return editor
}

function getLearnMorePath(sourcegraphBaseUrl: string): string {
    return new URL('/extensions/sourcegraph/open-in-editor', sourcegraphBaseUrl).href
}

function getUrlPattern(editor: Editor, editorSettings: EditorSettings): string {
    if (editor.urlPattern) {
        return editor.urlPattern
    }
    if (editor.id === 'vscode') {
        if (editorSettings.vscode?.useSSH) {
            if (!editorSettings.vscode.remoteHostForSSH) {
                throw new TypeError(
                    '`openineditor.vscode.mode` is set to "ssh" but `openineditor.vscode.remoteHostForSSH` is not set.'
                )
            }
            return `vscode://vscode-remote/ssh-remote+${editorSettings.vscode.remoteHostForSSH}%file:%line:%col`
        }
        return `${editorSettings.vscode?.useInsiders ? 'vscode-insiders' : 'vscode'}://file%file:%line:%col`
    }
    if (editor.isJetBrainsProduct) {
        if (editorSettings.jetbrains?.forceApi === 'builtInServer') {
            // Open files with IntelliJ's built-in REST API (port 63342) if useBuiltin is enabled instead of the idea:// protocol handler
            // ref: https://www.jetbrains.com/help/idea/php-built-in-web-server.html#configuring-built-in-web-server
            return 'http://localhost:63342/api/file%file:%line:%col'
        }
        return `${editor.id}://open?file=%file&line=%line&column=%col`
    }
    if (editor.id === 'custom') {
        return editorSettings.custom?.urlPattern ?? ''
    }
    throw new TypeError(`No url pattern found for editor ${editor.id}`)
}

function doReplacements(urlWithoutReplacements: string, replacements: EditorReplacements | undefined): string {
    let url = urlWithoutReplacements
    if (replacements) {
        for (const [search, replacement] of Object.keys(replacements)) {
            url = url.replace(new RegExp(search), replacement)
        }
    }
    return url
}
