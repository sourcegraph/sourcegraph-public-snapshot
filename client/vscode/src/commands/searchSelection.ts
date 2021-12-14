import vscode from 'vscode'

/**
 * Open active file in the browser on the configured Sourcegraph instance.
 *
 * TODO: implement opening remote Sourcegraph files. For now, just open local files in Sourcegraph.
 */
export async function searchSelection(): Promise<void> {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        throw new Error('No active editor')
    }
    const instanceUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
    const query = editor.document.getText(editor.selection)
    // check if the current file is a remote file or not
    if (query && typeof instanceUrl === 'string') {
        await openLinkInBrowser(
            `${instanceUrl}search?q=context:global+${encodeURIComponent(query)}&patternType=literal`
        )
    } else {
        await vscode.window.showInformationMessage('No selection detected.')
    }
}

// Open Link in Browser
export async function openLinkInBrowser(uri: string): Promise<void> {
    await vscode.env.openExternal(vscode.Uri.parse(uri))
}
