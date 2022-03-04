import vscode, { env } from 'vscode'

import { getSourcegraphFileUrl, repoInfo } from './git-helpers'
import { generateSourcegraphBlobLink, vsceUtms } from './initialize'
/**
 * Open active file in the browser on the configured Sourcegraph instance.
 */

export async function browserActions(action: string, logRedirectEvent: (uri: string) => void): Promise<void> {
    const editor = vscode.window.activeTextEditor
    if (!editor) {
        throw new Error('No active editor')
    }
    const uri = editor.document.uri
    let sourcegraphUrl = String()
    // check if the current file is a remote file or not
    if (uri.scheme === 'sourcegraph') {
        sourcegraphUrl = generateSourcegraphBlobLink(
            uri,
            editor.selection.start.line,
            editor.selection.start.character,
            editor.selection.end.line,
            editor.selection.end.character
        )
    } else {
        const repositoryInfo = await repoInfo(editor.document.uri.fsPath)
        if (!repositoryInfo) {
            return
        }
        const { remoteURL, branch, fileRelative } = repositoryInfo
        const instanceUrl = vscode.workspace.getConfiguration('sourcegraph').get('url')
        if (typeof instanceUrl === 'string') {
            // construct sourcegraph url for current file
            sourcegraphUrl = getSourcegraphFileUrl(instanceUrl, remoteURL, branch, fileRelative, editor) + vsceUtms
        }
    }
    // Log redirect events
    logRedirectEvent(sourcegraphUrl)

    // Open in browser or Copy file link
    if (action === 'open' && sourcegraphUrl) {
        await vscode.env.openExternal(vscode.Uri.parse(sourcegraphUrl))
    } else if (action === 'copy' && sourcegraphUrl) {
        const decodedUri = decodeURIComponent(sourcegraphUrl)
        await env.clipboard.writeText(decodedUri).then(() => vscode.window.showInformationMessage('Copied!'))
    } else {
        throw new Error(`Failed to ${action} file link: invalid URL`)
    }
}
