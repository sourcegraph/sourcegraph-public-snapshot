import * as vscode from 'vscode'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import { ChatViewProvider } from './chat/view'
import { WSChatClient } from './chat/ws'
import { CodyCompletionItemProvider, WSCompletionsClient, fetchAndShowCompletions } from './completions'
import { EmbeddingsClient } from './embeddings-client'

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
export async function activate(context: vscode.ExtensionContext) {
	console.log('codebot extension activated')
	const settings = vscode.workspace.getConfiguration()
	const documentProvider = new CompletionsDocumentProvider()
	const history = new History()
	history.register(context)

	const serverAddr = settings.get('codebot.conf.serverEndpoint')
	if (!serverAddr) {
		throw new Error('need to set server endpoint')
	}

	const embeddingsAddr: string | undefined = settings.get('codebot.conf.embeddingsEndpoint')
	if (!embeddingsAddr) {
		throw new Error('need to set embeddings endpoint')
	}

	const codebaseID: string | undefined = settings.get('codebot.conf.codebaseID')
	if (!codebaseID) {
		throw new Error('need to set codebaseID')
	}

	const wsCompletionsClient = await WSCompletionsClient.new(`ws://${serverAddr}/completions`)
	const wsChatClient = await WSChatClient.new(`ws://${serverAddr}/chat`)
	const embeddingsClient = new EmbeddingsClient(embeddingsAddr, codebaseID)

	const chatProvider = new ChatViewProvider(context.extensionPath, wsChatClient, embeddingsClient)

	const executeRecipe = async (recipe: string) => {
		await vscode.commands.executeCommand('cody.chat.focus')
		return chatProvider.executeRecipe(recipe)
	}

	context.subscriptions.push(
		vscode.workspace.registerTextDocumentContentProvider('codegen', documentProvider),
		vscode.languages.registerHoverProvider({ scheme: 'codegen' }, documentProvider),

		vscode.commands.registerCommand('vscode-codegen.ai-suggest', async () => {
			await fetchAndShowCompletions(wsCompletionsClient, documentProvider, history)
		}),

		vscode.commands.registerCommand('cody.recipe.explain-code', async () => executeRecipe('explainCode')),

		vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () =>
			executeRecipe('explainCodeHighLevel')
		),

		vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () =>
			executeRecipe('generateUnitTest')
		),

		vscode.window.registerWebviewViewProvider('cody.chat', chatProvider)
	)
}

// This method is called when your extension is deactivated
export function deactivate() {}
