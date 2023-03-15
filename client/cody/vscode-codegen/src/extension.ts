import * as vscode from 'vscode'

import { CommandsProvider } from './command/CommandsProvider'
import { ExtensionApi } from './extension-api'

export function activate(context: vscode.ExtensionContext): Promise<ExtensionApi> {
    console.log('Cody extension activated')

    // Register commands and webview
    return CommandsProvider(context)
}
