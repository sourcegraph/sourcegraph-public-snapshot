import vscode, { env } from 'vscode'

import { getSourcegraphFileUrl, repoInfo } from './helpers'

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
    let SourcegraphUrl
    // check if the current file is a remote file or not
    if (uri.scheme === 'sourcegraph') {
        SourcegraphUrl = `${encodeURI(uri.toString()).replace(uri.scheme, 'https')}?L${encodeURIComponent(
            String(editor.selection.start.line)
        )}:${encodeURIComponent(String(editor.selection.start.character))}-${encodeURIComponent(
            String(editor.selection.end.line)
        )}:${encodeURIComponent(String(editor.selection.end.character))}`
    } else {
        const repositoryInfo = await repoInfo(editor.document.uri.fsPath)
        if (!repositoryInfo) {
            return
        }
        const { remoteURL, branch, fileRelative } = repositoryInfo
        const instanceUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
        if (typeof instanceUrl === 'string') {
            // construct sourcegraph url for current file
            SourcegraphUrl = getSourcegraphFileUrl(instanceUrl, remoteURL, branch, fileRelative, editor)
        }
    }

    // Open in browser or Copy file link
    if (action === 'open' && SourcegraphUrl) {
        await vscode.env.openExternal(vscode.Uri.parse(SourcegraphUrl))
    } else if (action === 'copy' && SourcegraphUrl) {
        await env.clipboard
            .writeText(decodeURIComponent(SourcegraphUrl))
            .then(() => vscode.window.showInformationMessage('Copied!'))
    } else {
        throw new Error(`Failed to ${action} file link: invalid URL`)
    }
}
