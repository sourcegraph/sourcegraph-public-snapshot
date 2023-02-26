import * as vscode from 'vscode'

import {
	Completion,
	CompletionsArgs,
	LLMDebugInfo,
	WSCompletionResponse,
	WSCompletionsRequest,
} from '@sourcegraph/cody-common'

import { getReferences } from './autocomplete/completion-provider'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import { WSClient } from './wsclient'

interface CompletionCallbacks {
	onCompletions(completions: Completion[], debugInfo?: LLMDebugInfo): void
	onMetadata(metadata: any): void
	onDone(): void
	onError(err: string): void
}

export class WSCompletionsClient {
	static async new(addr: string, accessToken: string): Promise<WSCompletionsClient | null> {
		const wsclient = await WSClient.new<Omit<WSCompletionsRequest, 'requestId'>, WSCompletionResponse>(
			addr,
			accessToken
		)
		if (!wsclient) {
			return null
		}
		return new WSCompletionsClient(wsclient)
	}

	private constructor(private wsclient: WSClient<Omit<WSCompletionsRequest, 'requestId'>, WSCompletionResponse>) {}

	async getCompletions(args: CompletionsArgs, callbacks: CompletionCallbacks): Promise<void> {
		await this.wsclient.sendRequest(
			{
				kind: 'getCompletions',
				args,
			},
			resp => {
				try {
					switch (resp.kind) {
						case 'completion':
							callbacks.onCompletions(resp.completions, resp.debugInfo)
							return false
						case 'metadata':
							callbacks.onMetadata(resp.metadata)
							return false
						case 'error':
							callbacks.onError(resp.error)
							return false
						case 'done':
							callbacks.onDone()
							return true
						default:
							return false
					}
				} catch (error: any) {
					vscode.window.showErrorMessage(error)
					return false
				}
			}
		)
	}
}

async function getCompletionsArgs(
	history: History,
	document: vscode.TextDocument,
	position: vscode.Position
): Promise<CompletionsArgs> {
	const prefixRange = new vscode.Range(0, 0, position.line, position.character)
	const prefix = document.getText(prefixRange)
	const historyInfo = await history.getInfo()
	const references = await getReferences(document, position, [new vscode.Location(document.uri, prefixRange)])
	return {
		history: historyInfo,
		prefix,
		references,
		uri: document.uri.toString(),
	}
}

export async function fetchAndShowCompletions(
	wsclientPromise: Promise<WSCompletionsClient | null>,
	documentProvider: CompletionsDocumentProvider,
	history: History
) {
	const wsclient = await wsclientPromise
	if (!wsclient) {
		return
	}

	const currentEditor = vscode.window.activeTextEditor
	if (!currentEditor || currentEditor?.document.uri.scheme === 'codegen') {
		return
	}
	const filename = currentEditor.document.fileName
	const ext = filename.split('.').pop() || ''
	const completionsUri = vscode.Uri.parse('codegen:Completions.md')
	documentProvider.clearCompletions(completionsUri)
	vscode.workspace.openTextDocument(completionsUri).then(doc => {
		vscode.window.showTextDocument(doc, {
			preview: false,
			viewColumn: 2,
		})
	})

	try {
		const currentEditor = vscode.window.activeTextEditor
		if (!currentEditor) {
			throw new Error('no current active editor')
		}
		await wsclient.getCompletions(
			await getCompletionsArgs(history, currentEditor.document, currentEditor.selection.active),
			{
				onCompletions (completions: Completion[], debug?: LLMDebugInfo | undefined): void {
					documentProvider.addCompletions(completionsUri, ext, completions, debug)
				},
				onMetadata (metadata: any): void {
					console.log(`received metadata ${metadata}`)
				},
				onDone (): void {
					console.log('received done')
					documentProvider.setCompletionsDone(completionsUri)
				},
				onError (err: string): void {
					console.error(`received error ${err}`)
				},
			}
		)
	} catch (error: any) {
		vscode.window.showErrorMessage(error)
	}
}

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
	constructor(private wsclient: WSCompletionsClient, private history: History) {}

	async provideInlineCompletionItems(
		document: vscode.TextDocument,
		position: vscode.Position,
		context: vscode.InlineCompletionContext,
		token: vscode.CancellationToken
	): Promise<vscode.InlineCompletionItem[]> {
		// debounce
		await new Promise(resolve => setTimeout(resolve, 2000))
		if (token.isCancellationRequested) {
			console.log('cancelled')
			return []
		}

		return new Promise<vscode.InlineCompletionItem[]>(async (resolve, reject) => {
			const allCompletions: Completion[] = []
			await this.wsclient.getCompletions(await getCompletionsArgs(this.history, document, position), {
				onCompletions (completions: Completion[], debug?: LLMDebugInfo | undefined): void {
					allCompletions.push(
						...completions.map(c => ({
							// Limit inline completions to one line for now
							...c,
							insertText: c.insertText.slice(0, Math.max(0, c.insertText.indexOf('\n'))),
						}))
					)
				},
				onMetadata (metadata: any): void {
					console.log(`received metadata ${metadata}`)
				},
				onDone (): void {
					resolve(allCompletions.map(c => new vscode.InlineCompletionItem(c.insertText)))
				},
				onError (err: string): void {
					reject(`CodyComplemtionItemProvider: error fetching completions: ${err}`)
				},
			})
		})
	}
}
