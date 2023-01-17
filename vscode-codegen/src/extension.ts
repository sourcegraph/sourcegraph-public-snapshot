import * as vscode from 'vscode'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import { ChatViewProvider } from './chat/view'
import { WSChatClient } from './chat/ws'
import { CodyCompletionItemProvider, WSCompletionsClient, fetchAndShowCompletions } from './completions'
import { EmbeddingsClient } from './embeddings-client'
import { explainCode, explainCodeHighLevel, generateTest } from './command/testgen'
import { Message } from '@sourcegraph/cody-common'

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

	context.subscriptions.push(
		vscode.workspace.registerTextDocumentContentProvider('codegen', documentProvider),
		vscode.languages.registerHoverProvider({ scheme: 'codegen' }, documentProvider),

		vscode.commands.registerCommand('vscode-codegen.ai-suggest', async () => {
			await fetchAndShowCompletions(wsCompletionsClient, documentProvider, history)
		}),

		// TODO(beyang): rewrite this to be a property of the chat provider and the command invokes chatProvider.executeRecipe
		vscode.commands.registerCommand('codebot.generate-test', async (callback: (userMessage: string) => void) => {
			const userMessage = await generateTest(documentProvider)
			if (userMessage === null) {
				return
			}
			callback(userMessage)
		}),
		vscode.commands.registerCommand('codebot.explain-code', async (callback: (userMessage: string) => void) => {
			const userMessage = await explainCode()
			if (userMessage === null) {
				return
			}
			callback(userMessage)
		}),
		vscode.commands.registerCommand(
			'codebot.explain-code-high-level',
			async (callback: (userMessage: string) => void) => {
				const userMessage = await explainCodeHighLevel()
				if (userMessage === null) {
					return
				}
				callback(userMessage)
			}
		),

		// register inline completion provider
		vscode.languages.registerInlineCompletionItemProvider(
			{ pattern: '**' },
			new CodyCompletionItemProvider(wsCompletionsClient, history)
		),

		vscode.window.registerWebviewViewProvider('cody.chat', chatProvider)
	)
}

// This method is called when your extension is deactivated
export function deactivate() {}
