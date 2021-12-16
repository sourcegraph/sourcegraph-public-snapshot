import vscode from 'vscode'

/**
 * Open new search of the selected text in the browser on the configured Sourcegraph instance.
 *
 * TODO: implement opening new search within VSCE
 */
export async function searchSelection(): Promise<void> {
    const instanceUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
    const query = getSelectedText()
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

export function getSelectedText(): string {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        throw new Error('No active editor')
    }
    return editor.document.getText(editor.selection)
}
