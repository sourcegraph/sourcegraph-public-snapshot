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
    } else {
        await vscode.window.showInformationMessage('Local files are currently not supported')
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

export async function openLinkInBrowser(uri: string): Promise<void> {
    await vscode.env.openExternal(vscode.Uri.parse(uri))
}

export async function copyToClipboard(data: string): Promise<void> {
    await env.clipboard.writeText(data).then(() => vscode.window.showInformationMessage('Copied!'))
}
