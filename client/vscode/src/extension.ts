import vscode from 'vscode'

import { areExtensionsSame } from '@sourcegraph/shared/src/extensions/extensions'

export function activate(context: vscode.ExtensionContext): void {
    context.subscriptions.push(
        vscode.commands.registerCommand('sourcegraph.test', () => {
            const shouldDownloadExtensions = !areExtensionsSame([{ id: 'old' }], [{ id: 'new' }])
            console.log({ shouldDownloadExtensions })
        })
    )
}
