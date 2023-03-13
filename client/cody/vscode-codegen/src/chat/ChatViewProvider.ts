import fetch from 'node-fetch'
import * as vscode from 'vscode'

import { ChatMetadata, Feedback, Message } from '@sourcegraph/cody-common'

import { CODY_ACCESS_TOKEN_SECRET, updateConfiguration } from '../configuration'
import { EmbeddingsClient } from '../embeddings-client'
import { getRgPath } from '../rg'
import { TestSupport } from '../test-support'

import { renderMarkdown } from './markdown'
import { Transcript } from './prompt'
import { WSChatClient } from './ws'

export interface ChatMessage extends Omit<Message, 'text'> {
	displayText: string
	timestamp: string
	contextFiles?: string[]
}

// If the bot message ends with some prefix of the `Human:` stop
// sequence, trim if from the end.
const STOP_SEQUENCE_REGEXP = /(H|Hu|Hum|Huma|Human|Human:)$/

export class ChatViewProvider implements vscode.WebviewViewProvider {
	private transcript: ChatMessage[] = []
	private messageInProgress: ChatMessage | null = null

	private closeConnectionInProgressPromise: Promise<() => void> | null = null
	private prompt: Transcript
	private webview?: vscode.Webview

	private wsclient: Promise<WSChatClient | null>
	private embeddingsClient: EmbeddingsClient | null
	private rgPath = 'rg'

	private mode: 'development' | 'production'
	private tosVersion = 0

	constructor(
		private codebase: string,
		private extensionPath: string,
		private serverUrl: string,
		private accessToken: string,
		private embeddingsEndpoint: string,
		private contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
		private debug: boolean,
		private secretStorage: vscode.SecretStorage
	) {
		if (TestSupport.instance) {
			TestSupport.instance.chatViewProvider.set(this)
		}

		this.mode = this.debug ? 'development' : 'production'

		this.findRgPath()
			.then(path => (this.rgPath = path))
			.catch(error => console.error(error))

		this.wsclient = this.makeWSChatClient()
		this.embeddingsClient = this.makeEmbeddingClient()

		this.prompt = new Transcript(
			this.embeddingsClient,
			this.contextType,
			this.serverUrl,
			this.accessToken,
			this.rgPath
		)
	}

	private async findRgPath(): Promise<string> {
		return getRgPath(this.extensionPath)
	}

	private async onDidReceiveMessage(message: any, webview: vscode.Webview): Promise<void> {
		switch (message.command) {
			case 'initialized':
				await Promise.all([
					this.sendTranscript(),
					webview?.postMessage({
						type: 'token',
						value: await this.getAccessToken(),
						mode: this.mode,
					}),
				])
				break
			case 'reset':
				await this.onResetChat()
				break
			case 'submit':
				await this.onHumanMessageSubmitted(message.text)
				break
			case 'executeRecipe':
				await vscode.commands.executeCommand('cody.chat.focus')
				await this.executeRecipe(message.recipe)
				break
			case 'feedback':
				await this.sendFeedback(message.feedback)
				break
			case 'acceptTOS':
				await this.acceptTOS(message.version)
				break
			case 'setToken':
				await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.value)
				this.accessToken = message.value as string
				break
			case 'settings':
				await updateConfiguration('serverEndpoint', message.serverURL)
				await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.accessToken)
				this.serverUrl = message.value as string
				this.accessToken = message.accessToken as string
				if (this.accessToken && this.serverUrl) {
					this.wsclient = this.makeWSChatClient()
				}
				break
			case 'removeToken':
				await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
				break
			default:
				console.error('Invalid request type from Webview')
		}
	}

	private async getAccessToken(): Promise<string> {
		if (this.accessToken) {
			return this.accessToken
		}
		const token = await this.secretStorage.get(CODY_ACCESS_TOKEN_SECRET)
		return token || ''
	}

	private async acceptTOS(version: number): Promise<void> {
		this.tosVersion = version
		await vscode.commands.executeCommand('cody.accept-tos', version)
	}

	public async makeWSChatClient(): Promise<WSChatClient | null> {
		try {
			if (this.accessToken && this.serverUrl) {
				const wsclient = await WSChatClient.new(`${this.serverUrl}/chat`, this.accessToken)
				return wsclient
			}
			return null
		} catch (error) {
			console.error(error)
			return null
		}
	}

	private makeEmbeddingClient(): EmbeddingsClient | null {
		if (this.codebase && this.accessToken && this.embeddingsEndpoint && this.contextType === 'embeddings') {
			const embeddingsClient = new EmbeddingsClient(this.embeddingsEndpoint, this.accessToken, this.codebase)

			if (!embeddingsClient && this.contextType === 'embeddings') {
				// eslint-disable-next-line @typescript-eslint/no-floating-promises
				vscode.window.showInformationMessage(
					'Embeddings were not available (is `cody.codebase` set?), falling back to keyword context'
				)
				this.contextType = 'keyword'
			}

			return embeddingsClient
		}
		return null
	}

	private async sendFeedback(feedback: Feedback): Promise<void> {
		feedback.user = 'unknown'
		feedback.displayMessages = this.prompt.getDisplayMessages()
		feedback.transcript = this.prompt.getTranscript()
		feedback.feedbackVersion = 'v0'
		const resp = await fetch(`${this.serverUrl}/feedback`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: 'Bearer ' + this.accessToken,
			},
			body: JSON.stringify(feedback),
		})
		await resp.json()
	}

	private async sendPrompt(promptMessages: Message[], metadata: ChatMetadata, responsePrefix = ''): Promise<void> {
		const wsclient = await this.wsclient
		if (!wsclient) {
			return
		}

		await this.closeConnectionInProgress()
		await this.logSendPrompt(promptMessages)

		this.closeConnectionInProgressPromise = wsclient.chat(promptMessages, metadata, {
			onChange: text => this.onBotMessageChange(this.reformatBotMessage(text, responsePrefix)),
			onComplete: text => {
				const botMessage = this.reformatBotMessage(text, responsePrefix)
				this.logReceivedBotResponse(botMessage).catch(error => console.error(error))
				this.onBotMessageComplete(botMessage).catch(error => console.error(error))
			},
			onError: err => {
				void vscode.window.showErrorMessage(err)
			},
		})
	}

	private async logReceivedBotResponse(response: string): Promise<void> {
		await this.webview?.postMessage({
			type: 'debug',
			message: `RESPONSE (${response.length} characters):\n${response}`,
		})
	}

	private async logSendPrompt(promptMessages: Message[]): Promise<void> {
		const promptStr = promptMessages.map(msg => `${msg.speaker}: ${msg.text}`).join('\n\n')
		const debugMessage = `REQUEST (${promptStr.length} characters):\n` + promptStr
		await this.webview?.postMessage({ type: 'debug', message: debugMessage })
	}

	private async closeConnectionInProgress(): Promise<void> {
		if (!this.closeConnectionInProgressPromise) {
			return
		}
		const closeConnection = await this.closeConnectionInProgressPromise
		closeConnection()
		this.closeConnectionInProgressPromise = null
	}

	private async onResetChat(): Promise<void> {
		await this.closeConnectionInProgress()
		this.messageInProgress = null
		this.transcript = []
		this.prompt.reset()
		await this.sendTranscript()
	}

	private async onNewMessageSubmitted(text: string): Promise<void> {
		this.messageInProgress = {
			speaker: 'bot',
			displayText: '',
			timestamp: getShortTimestamp(),
		}

		this.transcript.push({
			speaker: 'you',
			displayText: renderMarkdown(text),
			timestamp: getShortTimestamp(),
		})

		await this.sendTranscript()
	}

	private async onHumanMessageSubmitted(text: string): Promise<void> {
		if (this.messageInProgress) {
			return
		}
		await this.onNewMessageSubmitted(text)
		const prompt = await this.prompt.addHumanMessage(text)
		await this.sendPrompt(prompt, { displayMessageLength: this.prompt.getDisplayMessages().length, recipeId: null })
	}

	public async executeRecipe(recipeId: string): Promise<void> {
		if (this.messageInProgress) {
			await vscode.window.showErrorMessage(
				'Cannot execute multiple recipes. Please wait for the current recipe to finish.'
			)
		}

		const messageInfo = await this.prompt.resetToRecipe(recipeId)
		if (!messageInfo) {
			console.error('unrecognized recipe prompt:', recipeId)
			return
		}
		const { display, prompt, botResponsePrefix } = messageInfo

		await this.showTab('ask')

		this.messageInProgress = {
			speaker: 'bot',
			displayText: '',
			timestamp: getShortTimestamp(),
		}
		this.transcript.push(
			...display.map(({ speaker, text }) => ({ speaker, displayText: text, timestamp: getShortTimestamp() }))
		)

		await this.sendTranscript()

		return this.sendPrompt(
			prompt,
			{ displayMessageLength: this.prompt.getDisplayMessages().length, recipeId },
			botResponsePrefix
		)
	}

	private reformatBotMessage(text: string, prefix: string): string {
		let reformattedMessage = prefix + text.trimEnd()

		const stopSequenceMatch = reformattedMessage.match(STOP_SEQUENCE_REGEXP)
		if (stopSequenceMatch) {
			reformattedMessage = reformattedMessage.slice(0, stopSequenceMatch.index)
		}
		// TODO: Detect if bot sent unformatted code without a markdown block.
		return fixOpenMarkdownCodeBlock(reformattedMessage)
	}

	private onBotMessageChange(text: string): void {
		this.messageInProgress = {
			speaker: 'bot',
			displayText: renderMarkdown(text),
			timestamp: getShortTimestamp(),
			contextFiles: this.prompt.getLastContextFiles(),
		}

		this.sendTranscript().catch(error => console.error(error))
	}

	private async onBotMessageComplete(text: string): Promise<void> {
		this.messageInProgress = null
		this.closeConnectionInProgressPromise = null
		this.transcript.push({
			speaker: 'bot',
			displayText: renderMarkdown(text),
			timestamp: getShortTimestamp(),
			contextFiles: this.prompt.getLastContextFiles(),
		})

		this.prompt.addBotMessage(text)

		await this.sendTranscript()
	}

	private async showTab(tab: string): Promise<void> {
		await this.webview?.postMessage({ type: 'showTab', tab })
	}

	private async sendTranscript(): Promise<void> {
		await this.webview?.postMessage({
			type: 'transcript',
			messages: this.transcript,
			messageInProgress: this.messageInProgress,
		})
	}

	public async resolveWebviewView(
		webviewView: vscode.WebviewView,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_context: vscode.WebviewViewResolveContext<unknown>,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		_token: vscode.CancellationToken
	): Promise<void> {
		this.webview = webviewView.webview
		const extensionPath = vscode.Uri.parse(this.extensionPath)
		const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')

		webviewView.webview.options = {
			enableScripts: true,
			localResourceRoots: [webviewPath],
		}

		// await vscode.commands.executeCommand('cody.get-accepted-tos-version')

		//   Create Webview
		const root = vscode.Uri.joinPath(webviewPath, 'index.html')
		const bytes = await vscode.workspace.fs.readFile(root)
		const decoded = new TextDecoder('utf-8').decode(bytes)
		const resources = webviewView.webview.asWebviewUri(webviewPath)

		const nonce = getNonce()

		webviewView.webview.html = decoded
			.replaceAll('/kodj/', `${resources.toString()}/`)
			.replace('/nonce/', nonce)
			.replace('/tos-version/', this.tosVersion.toString())

		webviewView.webview.onDidReceiveMessage(message => this.onDidReceiveMessage(message, webviewView.webview))
	}

	public transcriptForTesting(testing: TestSupport): ChatMessage[] {
		if (!testing) {
			console.error('used ForTesting method without test support object')
			return []
		}
		return this.transcript
	}

	public async configChangeDetected(changes: string): Promise<void> {
		switch (changes) {
			case 'token':
				await vscode.window.showInformationMessage('Token has been updated.')
				// TODO: Reset clients.
				// Notify user with exisiting conversations to reload VS Code.
				// await vscode.window.showWarningMessage('Please reload VS Code to use the new token.')
				break
			default:
				await vscode.window.showInformationMessage('Your configurations for Cody have been updated.')
		}
	}
}

function fixOpenMarkdownCodeBlock(text: string): string {
	const occurances = text.split('```').length - 1
	if (occurances % 2 === 1) {
		return text + '\n```'
	}
	return text
}

function padTimePart(timePart: number): string {
	return timePart < 10 ? `0${timePart}` : timePart.toString()
}

function getShortTimestamp(): string {
	const date = new Date()
	return `${padTimePart(date.getHours())}:${padTimePart(date.getMinutes())}`
}

function getNonce(): string {
	let text = ''
	const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
	for (let i = 0; i < 32; i++) {
		text += possible.charAt(Math.floor(Math.random() * possible.length))
	}
	return text
}
