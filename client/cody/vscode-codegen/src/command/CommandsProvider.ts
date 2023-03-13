import * as vscode from 'vscode'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { CODY_ACCESS_TOKEN_SECRET, ConfigurationUseContext, getConfiguration } from '../configuration'

// Registers Commands and Webview at extension start up
export const CommandsProvider = async (context: vscode.ExtensionContext): Promise<void> => {
	const config = getConfiguration(vscode.workspace.getConfiguration())
	const accessToken = (await context.secrets.get(CODY_ACCESS_TOKEN_SECRET)) || ''
	const useContext: ConfigurationUseContext = config.useContext

	// Create chat webview
	const chatProvider = new ChatViewProvider(
		config.codebase || '',
		context.extensionPath,
		config.serverEndpoint,
		accessToken,
		config.embeddingsEndpoint,
		useContext,
		config.debug,
		context.secrets
	)

	vscode.window.registerWebviewViewProvider('cody.chat', chatProvider)

	await vscode.commands.executeCommand('setContext', 'sourcegraph.cody.activated', true)

	const disposables: vscode.Disposable[] = []

	disposables.push(
		// Toggle Chat
		vscode.commands.registerCommand('sourcegraph.cody.toggleEnabled', async () => {
			const config = vscode.workspace.getConfiguration()
			await config.update(
				'sourcegraph.cody.enable',
				!config.get('sourcegraph.cody.enable'),
				vscode.ConfigurationTarget.Global
			)
		}),
		// Access token
		vscode.commands.registerCommand('cody.set-access-token', async (args: any[]) => {
			const tokenInput = args?.length ? (args[0] as string) : await vscode.window.showInputBox()
			if (tokenInput === undefined || tokenInput === '') {
				return
			}
			await context.secrets.store(CODY_ACCESS_TOKEN_SECRET, tokenInput)
		}),
		vscode.commands.registerCommand('cody.delete-access-token', async () =>
			context.secrets.delete(CODY_ACCESS_TOKEN_SECRET)
		),
		// TOS
		vscode.commands.registerCommand('cody.accept-tos', async version => {
			if (typeof version !== 'number') {
				// eslint-disable-next-line @typescript-eslint/restrict-template-expressions
				void vscode.window.showErrorMessage(`TOS version was not a number: ${version}`)
				return
			}
			await context.globalState.update('cody.tos-version-accepted', version)
		}),
		vscode.commands.registerCommand('cody.get-accepted-tos-version', async () => {
			const version = await context.globalState.get('cody.tos-version-accepted')
			return version
		}),
		// Commands
		vscode.commands.registerCommand('cody.recipe.explain-code', async () => executeRecipe('explainCode')),
		vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () =>
			executeRecipe('explainCodeHighLevel')
		),
		vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () =>
			executeRecipe('generateUnitTest')
		),
		vscode.commands.registerCommand('cody.recipe.generate-docstring', async () =>
			executeRecipe('generateDocstring')
		),
		vscode.commands.registerCommand('cody.recipe.translate-to-language', async () =>
			executeRecipe('translateToLanguage')
		),
		vscode.commands.registerCommand('cody.recipe.git-history', async () => executeRecipe('gitHistory'))
	)

	// Watch all relevant configuration and secrets for changes.
	context.subscriptions.push(
		vscode.workspace.onDidChangeConfiguration(async event => {
			if (event.affectsConfiguration('cody') || event.affectsConfiguration('sourcegraph')) {
				await chatProvider.configChangeDetected('endpoints')
			}
		})
	)
	context.subscriptions.push(
		context.secrets.onDidChange(async event => {
			if (event.key === CODY_ACCESS_TOKEN_SECRET) {
				await chatProvider.configChangeDetected('token')
			}
		})
	)

	const executeRecipe = async (recipe: string): Promise<void> => {
		await vscode.commands.executeCommand('cody.chat.focus')
		await chatProvider.executeRecipe(recipe)
	}

	vscode.Disposable.from(...disposables)
}
