import * as vscode from 'vscode'

import { CommandsProvider } from './command/CommandsProvider'

export async function activate(context: vscode.ExtensionContext): Promise<void> {
	console.log('Cody extension activated')

	// Register commands and webview
	await CommandsProvider(context)
}
