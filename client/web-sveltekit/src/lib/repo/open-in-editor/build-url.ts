import { type EditorSettings, getEditor, supportedEditors, isProjectPathValid, getProjectPath } from '$lib/web'

/**
 * This function is a copy of client/web/src/open-in-editor/build-util.ts#getEditorSettingsErrorMessage with
 * wording adjusted to not contain markdown.
 */
export function getEditorSettingsErrorMessage(editorSettings: EditorSettings | undefined): string | undefined {
    if (!editorSettings) {
        return 'Add `openInEditor` to your user settings to open files in the editor. Click to learn more.'
    }

    const projectPath = getProjectPath(editorSettings)

    if (typeof projectPath !== 'string') {
        return 'Add `projectPaths.default` or some OS-specific path to your user settings to open files in the editor. Click to learn more.'
    }

    // Skip this check on Windows because path.isAbsolute only checks Linux and macOS compatible paths reliably
    if (!isProjectPathValid(projectPath)) {
        return `\`projectPaths.default\` (or your current OS-specific setting) \`${projectPath}\` is not an absolute path. Please correct the error in your user settings.`
    }

    if (!editorSettings.editorIds?.length) {
        return 'Add `editorIds` to your user settings to open files. Click to learn more.'
    }
    const validEditorCount = editorSettings.editorIds.map(id => getEditor(id)).filter(editor => editor).length

    if (validEditorCount !== editorSettings.editorIds.length) {
        return (
            'Setting `editorIds` must be set to a valid array of values in your user settings to open files. Supported editors: ' +
            supportedEditors.map(editor => editor.id).join(', ')
        )
    }
    if (editorSettings.editorIds?.includes('custom') && typeof editorSettings['custom.urlPattern'] !== 'string') {
        return 'Add `custom.urlPattern` to your user settings for custom editor to open files. Click to learn more.'
    }

    if (editorSettings['vscode.useSSH'] && !editorSettings['vscode.remoteHostForSSH']) {
        return '`vscode.useSSH` is set to "true" but `vscode.remoteHostForSSH` is not set.'
    }

    return undefined
}
