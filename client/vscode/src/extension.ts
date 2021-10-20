import 'cross-fetch/polyfill'
import vscode from 'vscode'

import { requestGraphQLFromVSCode } from './backend/requestGraphQl'
import { SourcegraphVSCodeExtensionAPI } from './webview/contract'
import { initializeWebview } from './webview/initialize'

export function activate(context: vscode.ExtensionContext): void {
    // TODO from extension settings
    const SOURCEGRAPH_URL = 'https://sourcegraph.com'
    // TODO access token

    const sourcegraphVSCodeExtensionAPI: SourcegraphVSCodeExtensionAPI = {
        requestGraphQL: async (request, variables) => requestGraphQLFromVSCode(request, variables, SOURCEGRAPH_URL),
        ping: () => 'pong!',
    }

    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.test', () => {
            initializeWebview({
                extensionPath: context.extensionPath,
                id: 'sourcegraphSearch',
                title: 'Sourcegraph Search',
                route: 'search',
                sourcegraphVSCodeExtensionAPI,
            })
        })
    )
}
