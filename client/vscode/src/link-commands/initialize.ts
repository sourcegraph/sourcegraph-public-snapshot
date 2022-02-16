import vscode from 'vscode'

import { SourcegraphUri } from '../file-system/SourcegraphUri'

import { browserActions } from './browserActionsNode'

export function initializeCodeSharingCommands({ context }: { context: vscode.ExtensionContext }): void {
    // Open local file or remote Sourcegraph file in browser
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.openInBrowser', async () => {
            await browserActions('open')
        })
    )

    // Copy Sourcegraph link to file
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.copyFileLink', async () => {
            await browserActions('copy')
        })
    )

    // Search Selected Text in Sourcegraph Search Tab
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.selectionSearchWeb', async () => {
            const instanceUrl =
                vscode.workspace.getConfiguration('sourcegraph').get<string>('url') || 'https://sourcegraph.com'
            const editor = vscode.window.activeTextEditor
            const selectedQuery = editor?.document.getText(editor.selection)
            if (!editor || !selectedQuery) {
                throw new Error('No selection detected')
            }
            const uri = `${instanceUrl}/search?q=context:global+${encodeURIComponent(
                selectedQuery
            )}&patternType=literal${vsceUtms}`
            await vscode.env.openExternal(vscode.Uri.parse(uri))
        })
    )
}

export const vsceUtms =
    '&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=vsce-commands'

export function generateSourcegraphBlobLink(
    uri: vscode.Uri,
    startLine: number,
    startChar: number,
    endLine: number,
    endChar: number
): string {
    const instanceUrl = vscode.workspace.getConfiguration('sourcegraph').get<string>('url') || 'https://sourcegraph.com'
    // Using SourcegraphUri.parse to properly decode repo revision
    const decodedUri = SourcegraphUri.parse(uri.toString()).uri
    return `${decodedUri.replace(uri.scheme, instanceUrl.startsWith('https') ? 'https' : 'http')}?L${encodeURIComponent(
        String(startLine)
    )}:${encodeURIComponent(String(startChar))}-${encodeURIComponent(String(endLine))}:${encodeURIComponent(
        String(endChar)
    )}${vsceUtms}`
}
