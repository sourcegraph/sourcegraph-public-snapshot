import path from 'path'

import 'cross-fetch/polyfill'
import * as vscode from 'vscode'

import { EmbeddingsClient } from '../clients/embeddings-client'
import { WSChatClient } from '../clients/ws'
import { Feedback, Message } from '../types'

import { CODY_ACCESS_TOKEN_SECRET } from './configuration'
import { renderMarkdown } from './markdown'
import { Transcript } from './prompt'

export interface ChatMessage extends Omit<Message, 'text'> {
    displayText: string
    timestamp: string
    contextFiles?: string[]
}

// If the bot message ends with some prefix of the `Human:` stop
// sequence, trim if from the end.
const STOP_SEQUENCE_REGEXP = /(H|Hu|Hum|Huma|Human|Human:)$/

export class ChatViewProvider implements vscode.WebviewViewProvider {
    private readonly staticDir = path.join('resources', 'chat')
    private readonly staticFiles = {
        css: ['vscode.css', 'tabs.css', 'style.css', 'highlight.css'],
        js: ['tabs.js', 'index.js'],
    }

    private transcript: ChatMessage[] = []
    private messageInProgress: ChatMessage | null = null

    private closeConnectionInProgressPromise: Promise<() => void> | null = null
    private prompt: Transcript
    private webview?: vscode.Webview
    private wsclient: Promise<WSChatClient | null>

    private embeddingsClient: EmbeddingsClient | null

    constructor(
        private codebase: string,
        private extensionPath: string,
        private serverUrl: string,
        private accessToken: string,
        private embeddingsEndpoint: string,
        private contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        private debug: boolean,
        private rgPath: string,
        private secretStorage: vscode.SecretStorage
    ) {
        //TODO
        // if (TestSupport.instance) {
        //   TestSupport.instance.chatViewProvider.set(this);
        // }

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

    public async makeWSChatClient(): Promise<WSChatClient | null> {
        try {
            const wsclient = await WSChatClient.new(`${this.serverUrl}/chat`, this.accessToken)
            return wsclient
        } catch (error) {
            console.error(error)
            return null
        }
    }

    private makeEmbeddingClient() {
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
        } else {
            return null
        }
    }

    async resolveWebviewView(
        webviewView: vscode.WebviewView,
        _context: vscode.WebviewViewResolveContext<unknown>,
        _token: vscode.CancellationToken
    ): Promise<void> {
        // const tosVersion = await vscode.commands.executeCommand(
        //   'cody.get-accepted-tos-version'
        // );

        this.webview = webviewView.webview
        const extensionPath = vscode.Uri.parse(this.extensionPath)
        const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')

        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [webviewPath],
        }

        webviewView.webview.onDidReceiveMessage(message => this.onDidReceiveMessage(message, webviewView.webview))

        //   Create Webview
        const root = vscode.Uri.joinPath(webviewPath, 'index.html')
        const bytes = await vscode.workspace.fs.readFile(root)
        const decoded = new TextDecoder('utf-8').decode(bytes)
        const resources = webviewView.webview.asWebviewUri(webviewPath)

        webviewView.webview.html = decoded.replaceAll('/kodj', `${resources.toString()}/`)
    }

    private async onDidReceiveMessage(message: any, webview: vscode.Webview): Promise<void> {
        switch (message.command) {
            case 'initialized':
                this.sendTranscript()
                break
            case 'reset':
                this.onResetChat()
                break
            case 'submit':
                this.onHumanMessageSubmitted(message.text)
                break
            case 'executeRecipe':
                this.executeRecipe(message.recipe)
                break
            case 'feedback':
                this.sendFeedback(message.feedback)
                break
            case 'acceptTOS':
                this.acceptTOS(message.version)
                break
            case 'set-token':
                await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.value)
                break
            case 'settings':
                vscode.workspace.getConfiguration('cody').update('serverEndpoint', message.serverURL)
                this.serverUrl = message.value
                await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.accessToken)
                this.accessToken = message.accessToken
                if (this.accessToken && this.serverUrl) {
                    this.wsclient = this.makeWSChatClient()
                }
                break
            case 'remove-token':
                await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
                break
            case 'get-token':
                const accessToken = await this.secretStorage.get(CODY_ACCESS_TOKEN_SECRET)
                webview.postMessage({
                    type: 'auth',
                    value: accessToken || '',
                })
                if (accessToken) {
                    this.wsclient = this.makeWSChatClient()
                }
                break
            default:
                console.error('Invalid request type from Webview')
        }
    }

    private async acceptTOS(version: number) {
        vscode.commands.executeCommand('cody.accept-tos', version)
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

    private async sendPrompt(promptMessages: Message[], responsePrefix = ''): Promise<void> {
        const wsclient = await this.wsclient
        if (!wsclient) {
            return
        }

        await this.closeConnectionInProgress()
        await this.logSendPrompt(promptMessages)

        this.closeConnectionInProgressPromise = wsclient.chat(promptMessages, {
            onChange: text => this.onBotMessageChange(this.reformatBotMessage(text, responsePrefix)),
            onComplete: text => {
                const botMessage = this.reformatBotMessage(text, responsePrefix)
                this.logReceivedBotResponse(botMessage)
                this.onBotMessageComplete(botMessage)
            },
            onError: err => {
                vscode.window.showErrorMessage(err)
            },
        })
    }

    private logReceivedBotResponse(response: string) {
        this.webview?.postMessage({
            type: 'debug',
            message: `RESPONSE (${response.length} characters):\n${response}`,
        })
    }

    private async logSendPrompt(promptMessages: Message[]): Promise<void> {
        const promptStr = promptMessages.map(msg => `${msg.speaker}: ${msg.text}`).join('\n\n')
        const debugMessage = `REQUEST (${promptStr.length} characters):\n` + promptStr
        this.webview?.postMessage({ type: 'debug', message: debugMessage })
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
        this.sendTranscript()
    }

    private onNewMessageSubmitted(text: string): void {
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

        this.sendTranscript()
    }

    private async onHumanMessageSubmitted(text: string): Promise<void> {
        if (this.messageInProgress) {
            return
        }
        this.onNewMessageSubmitted(text)
        const prompt = await this.prompt.addHumanMessage(text)
        await this.sendPrompt(prompt)
    }

    async executeRecipe(recipeID: string): Promise<void> {
        if (this.messageInProgress) {
            vscode.window.showErrorMessage(
                'Cannot execute multiple recipes. Please wait for the current recipe to finish.'
            )
        }

        const messageInfo = await this.prompt.resetToRecipe(recipeID)
        if (!messageInfo) {
            console.error('unrecognized recipe prompt:', recipeID)
            return
        }
        const { display, prompt, botResponsePrefix } = messageInfo

        this.showTab('ask')

        this.messageInProgress = {
            speaker: 'bot',
            displayText: '',
            timestamp: getShortTimestamp(),
        }

        this.transcript.push(
            ...display.map(({ speaker, text }) => ({
                speaker,
                displayText: text,
                timestamp: getShortTimestamp(),
            }))
        )

        this.webview &&
            this.webview.postMessage({
                type: 'transcripts',
                messages: this.transcript,
                messageInProgress: this.messageInProgress,
            })

        return this.sendPrompt(prompt, botResponsePrefix)
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

        this.sendTranscript()
    }

    private onBotMessageComplete(text: string): void {
        this.messageInProgress = null
        this.closeConnectionInProgressPromise = null
        this.transcript.push({
            speaker: 'bot',
            displayText: renderMarkdown(text),
            timestamp: getShortTimestamp(),
            contextFiles: this.prompt.getLastContextFiles(),
        })

        this.prompt.addBotMessage(text)

        this.sendTranscript()

        this.webview &&
            this.webview.postMessage({
                type: 'transcripts',
                messages: this.transcript,
                messageInProgress: this.messageInProgress,
            })
    }

    private showTab(tab: string): void {
        this.webview?.postMessage({ type: 'showTab', tab })
    }

    private sendTranscript() {
        this.webview?.postMessage({
            type: 'transcript',
            messages: this.transcript,
            messageInProgress: this.messageInProgress,
        })
    }

    //TODO
    //   public transcriptForTesting(testing: TestSupport): ChatMessage[] {
    //     if (!testing) {
    //       console.error('used ForTesting method without test support object');
    //       return [];
    //     }
    //     return this.transcript;
    //   }
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

function getNonce() {
    let text = ''
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
    for (let i = 0; i < 32; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length))
    }
    return text
}
