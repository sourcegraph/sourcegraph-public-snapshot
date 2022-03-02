import vscode, { env } from 'vscode'

import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { version } from '../../package.json'
import { ExtensionCoreAPI } from '../contract'
import { ANONYMOUS_USER_ID_KEY, INSTANCE_VERSION_NUMBER_KEY } from '../settings/LocalStorageService'

import { getSourcegraphFileUrl, repoInfo } from './git-helpers'
import { checkEventSourceSupport, generateSourcegraphBlobLink, vsceUtms } from './initialize'
/**
 * Open active file in the browser on the configured Sourcegraph instance.
 */

export async function browserActions(action: string, extensionCoreAPI: ExtensionCoreAPI): Promise<void> {
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
    const instanceVersion = extensionCoreAPI.getLocalStorageItem(INSTANCE_VERSION_NUMBER_KEY)
    const userEventVariables = {
        event: 'IDERedirects',
        userCookieID: extensionCoreAPI.getLocalStorageItem(ANONYMOUS_USER_ID_KEY),
        referrer: 'VSCE',
        url: sourcegraphUrl,
        source: checkEventSourceSupport(instanceVersion) ? EventSource.IDEEXTENSION : EventSource.BACKEND,
        argument: JSON.stringify({ platform: 'vscode', version, action }),
    }
    extensionCoreAPI.logEvents(userEventVariables)
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
