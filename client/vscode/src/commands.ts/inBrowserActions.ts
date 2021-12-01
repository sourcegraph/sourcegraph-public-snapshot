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
    const repositoryInfo = await repoInfo(editor.document.uri.fsPath)
    if (!repositoryInfo) {
        return
    }
    const { remoteURL, branch, fileRelative } = repositoryInfo

    const SourcegraphUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')

    if (typeof SourcegraphUrl === 'string') {
        // construct sourcegraph url for current file
        const fileUri = getSourcegraphFileUrl(SourcegraphUrl, remoteURL, branch, fileRelative, editor)
        // Open in browser
        if (action === 'open') {
            await vscode.env.openExternal(vscode.Uri.parse(fileUri))
        } else if (action === 'copy') {
            await env.clipboard.writeText(fileUri).then(() => vscode.window.showInformationMessage('Copied!'))
        }
    }
}
