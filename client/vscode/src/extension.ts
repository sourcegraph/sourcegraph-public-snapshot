import vscode from 'vscode'

import { initializeWebview } from './webview/initialize'

export function activate(context: vscode.ExtensionContext): void {
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.test', () => {
            initializeWebview({
                extensionPath: context.extensionPath,
                id: 'sourcegraphSearch',
                title: 'Sourcegraph Search',
                route: 'search',
            })
        })
    )
}
