import vscode, { env } from 'vscode'

import { decodeUri } from './helpers'

/**
 * Open active file in the browser on the configured Sourcegraph instance.
 *
 * TODO: implement opening remote Sourcegraph files. For now, just open local files in Sourcegraph.
 */
export async function inBrowserActions(action: string): Promise<void> {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        throw new Error('No active editor')
    }
    const uri = editor.document.uri
    const instanceUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
    let sourcegraphUrl = String()
    // check if the current file is a remote file or not
    if (uri.scheme === 'sourcegraph') {
        // Using SourcegraphUri.parse to properly decode repo revision
        const decodedUri = decodeUri(uri.toString())
        sourcegraphUrl = `${decodedUri.replace(uri.scheme, 'https')}?L${encodeURIComponent(
            String(editor.selection.start.line)
        )}:${encodeURIComponent(String(editor.selection.start.character))}-${encodeURIComponent(
            String(editor.selection.end.line)
        )}:${encodeURIComponent(String(editor.selection.end.character))}`
    } else if (uri.authority === 'github' && typeof instanceUrl === 'string') {
        // For remote github files
        const repoInfo = uri.fsPath.split('/')
        const repositoryName = `${repoInfo[1]}/${repoInfo[2]}`
        const filePath = repoInfo.length === 4 ? repoInfo[3] : repoInfo.slice(3).join('/')
        sourcegraphUrl = `${instanceUrl}github.com/${repositoryName}/-/blob/${filePath || ''}`
    } else {
        await vscode.window.showInformationMessage('Local files are currently not supported on VS Code Web')
    }

    // Open in browser or Copy file link
    if (action === 'open' && sourcegraphUrl) {
        await openLinkInBrowser(sourcegraphUrl)
    } else if (action === 'copy' && sourcegraphUrl) {
        await copyToClipboard(decodeURIComponent(sourcegraphUrl))
    } else {
        throw new Error(`Failed to ${action} file link: invalid URL / not supported`)
    }
}

// Open Link in Browser
export async function openLinkInBrowser(uri: string): Promise<void> {
    await vscode.env.openExternal(vscode.Uri.parse(uri))
}

// Copy Link to Clipboard
export async function copyToClipboard(data: string): Promise<void> {
    await env.clipboard.writeText(data).then(() => vscode.window.showInformationMessage('Copied!'))
}
