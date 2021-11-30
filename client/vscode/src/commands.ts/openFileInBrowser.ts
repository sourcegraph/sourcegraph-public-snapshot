import vscode from 'vscode'

/**
 * Open active file in the browser on the configured Sourcegraph instance.
 *
 * TODO: implement opening remote Sourcegraph files. For now, just open local files in Sourcegraph.
 */
export function openFileInBrowser(): void {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        throw new Error('No active editor')
    }
    console.log({ editor })
}
