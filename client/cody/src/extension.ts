import * as vscode from 'vscode'

import { CommandsProvider } from './commands/CommandsProvider'

export async function activate(context: vscode.ExtensionContext): Promise<void> {
    console.log('Activating Cody')

    // set current workspace as codebase
    const workspace = vscode.workspace.workspaceFolders?.[0] ? vscode.workspace.workspaceFolders[0].uri : null

    if (workspace) {
        const currentCodebase = workspace.path.split('/').splice(-2).join('.')
        await vscode.workspace.getConfiguration('cody').update('codebase', currentCodebase)
    }

    // register commands and webview
    await CommandsProvider(context)
}

// This method is called when your extension is deactivated
export function deactivate(): void {}
