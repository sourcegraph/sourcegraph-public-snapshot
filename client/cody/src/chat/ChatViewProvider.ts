import * as vscode from 'vscode'

import { version as packageVersion } from '../../package.json'
import { ChatHistory } from '../../webviews/utils/types'
import { CodebaseContext } from '../codebase-context'
import { LocalStorage } from '../command/LocalStorageProvider'
import { CODY_ACCESS_TOKEN_SECRET, getAccessToken, SecretStorage } from '../command/secret-storage'
import { updateConfiguration } from '../configuration'
import { Editor } from '../editor'
import { VSCodeEditor } from '../editor/vscode-editor'
import { configureExternalServices } from '../external-services'
import { IntentDetector } from '../intent-detector'
import { getRgPath } from '../rg'
import { Message } from '../sourcegraph-api'
import { TestSupport } from '../test-support'

import { ChatClient } from './chat'
import { renderMarkdown } from './markdown'
import { getRecipe } from './recipes'
import { Transcript } from './transcript'
import { ChatMessage } from './transcript/messages'

// If the bot message ends with some prefix of the `Human:` stop
// sequence, trim if from the end.
const STOP_SEQUENCE_REGEXP = /(H|Hu|Hum|Huma|Human|Human:)$/

export class ChatViewProvider implements vscode.WebviewViewProvider {
    private isMessageInProgress = false
    private cancelCompletionCallback: (() => void) | null = null
    private webview?: vscode.Webview

    private tosVersion = packageVersion
    private editor: Editor

    private currentChatID: string = ''
    private inputHistory: string[] = []
    private chatHistory: ChatHistory = {}

    constructor(
        private extensionPath: string,
        private transcript: Transcript,
        private chat: ChatClient,
        private intentDetector: IntentDetector,
        private codebaseContext: CodebaseContext,
        private secretStorage: SecretStorage,
        private contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        private rgPath: string,
        private mode: 'development' | 'production',
        private localStorage: LocalStorage
    ) {
        if (TestSupport.instance) {
            TestSupport.instance.chatViewProvider.set(this)
        }
        this.editor = new VSCodeEditor()
        this.createNewChatID()
    }

    static async create(
        extensionPath: string,
        codebase: string,
        serverEndpoint: string,
        contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        debug: boolean,
        secretStorage: SecretStorage,
        localStorage: LocalStorage
    ): Promise<ChatViewProvider> {
        const mode = debug ? 'development' : 'production'
        const rgPath = await getRgPath(extensionPath)

        const { intentDetector, codebaseContext, chatClient } = await configureExternalServices(
            contextType,
            codebase,
            rgPath,
            serverEndpoint,
            secretStorage,
            mode
        )

        return new ChatViewProvider(
            extensionPath,
            new Transcript(),
            chatClient,
            intentDetector,
            codebaseContext,
            secretStorage,
            contextType,
            rgPath,
            mode,
            localStorage
        )
    }

    private async onDidReceiveMessage(message: any, webview: vscode.Webview): Promise<void> {
        switch (message.command) {
            case 'initialized':
                this.sendToken()
                this.sendTranscript()
                this.sendChatHistory()
                break
            case 'reset':
                this.onResetChat()
                this.sendChatHistory()
                break
            case 'submit':
                await this.onHumanMessageSubmitted(message.text)
                break
            case 'executeRecipe':
                await this.executeRecipe(message.recipe)
                break
            case 'acceptTOS':
                await this.acceptTOS(message.version)
                break
            case 'settings':
                await updateConfiguration('serverEndpoint', message.serverEndpoint)
                await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.accessToken)
                break
            case 'removeToken':
                await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
                break
            case 'removeHistory':
                await this.localStorage.removeChatHistory()
                break
            case 'links':
                await vscode.env.openExternal(vscode.Uri.parse(message.value))
                break
            default:
                console.error('Invalid request type from Webview')
        }
    }

    private async acceptTOS(version: string): Promise<void> {
        this.tosVersion = version
        await vscode.commands.executeCommand('cody.accept-tos', version)
    }

    private createNewChatID(): void {
        this.currentChatID = new Date(Date.now()).toUTCString()
    }

    private async sendPrompt(promptMessages: Message[], responsePrefix = ''): Promise<void> {
        this.cancelCompletion()

        this.cancelCompletionCallback = this.chat.chat(promptMessages, {
            onChange: text => this.onBotMessageChange(this.reformatBotMessage(text, responsePrefix)),
            onComplete: () => this.onBotMessageComplete(),
            onError: err => {
                vscode.window.showErrorMessage(err)
            },
        })
    }

    private cancelCompletion(): void {
        this.cancelCompletionCallback?.()
        this.cancelCompletionCallback = null
    }

    private onResetChat(): void {
        this.createNewChatID()
        this.cancelCompletion()
        this.isMessageInProgress = false
        this.transcript.reset()
        this.sendTranscript()
    }

    private async onHumanMessageSubmitted(text: string): Promise<void> {
        this.inputHistory.push(text)
        this.executeRecipe('chat-question', text)
    }

    public async executeRecipe(recipeId: string, humanChatInput: string = ''): Promise<void> {
        if (this.isMessageInProgress) {
            await vscode.window.showErrorMessage(
                'Cannot execute multiple recipes. Please wait for the current recipe to finish.'
            )
        }
        const recipe = getRecipe(recipeId)
        if (!recipe) {
            return
        }

        const interaction = await recipe.getInteraction(
            humanChatInput,
            this.editor,
            this.intentDetector,
            this.codebaseContext
        )
        if (!interaction) {
            return
        }
        this.isMessageInProgress = true
        this.transcript.addInteraction(interaction)

        this.showTab('chat')
        this.sendTranscript()

        const prompt = await this.transcript.toPrompt()
        this.sendPrompt(prompt, interaction.getAssistantMessage().prefix ?? '')
    }

    private reformatBotMessage(text: string, prefix: string): string {
        let reformattedMessage = prefix + text.trimEnd()

        const stopSequenceMatch = reformattedMessage.match(STOP_SEQUENCE_REGEXP)
        if (stopSequenceMatch) {
            reformattedMessage = reformattedMessage.slice(0, stopSequenceMatch.index)
        }
        // TODO: Detect if bot sent unformatted code without a markdown block.
        return this.fixOpenMarkdownCodeBlock(reformattedMessage)
    }

    private onBotMessageChange(text: string): void {
        this.transcript.addAssistantResponse(text, renderMarkdown(text))
        this.sendTranscript()
    }

    private async onBotMessageComplete(): Promise<void> {
        this.isMessageInProgress = false
        this.cancelCompletionCallback = null
        this.sendTranscript()
        this.saveChatHistory()
    }

    private async showTab(tab: string): Promise<void> {
        await vscode.commands.executeCommand('cody.chat.focus')
        await this.webview?.postMessage({ type: 'showTab', tab })
    }
    /**
     * Sends chat transcript to webview
     */
    private sendTranscript(): void {
        this.webview?.postMessage({
            type: 'transcript',
            messages: this.transcript.toChat(),
            isMessageInProgress: this.isMessageInProgress,
        })
    }
    /**
     * Sends access token to webview
     */
    private async sendToken(): Promise<void> {
        this.webview?.postMessage({
            type: 'token',
            value: await getAccessToken(this.secretStorage),
            mode: this.mode,
        })
    }
    /**
     * Save chat history
     */
    private async saveChatHistory(): Promise<void> {
        if (this.transcript) {
            this.chatHistory[this.currentChatID] = this.transcript.toChat()
            const userHistory = {
                chat: this.chatHistory,
                input: this.inputHistory,
            }
            await this.localStorage.setChatHistory(userHistory)
        }
    }
    /**
     * Sends chat history to webview
     */
    private sendChatHistory(): void {
        const localHistory = this.localStorage.getChatHistory()
        if (localHistory) {
            this.chatHistory = localHistory.chat
            this.inputHistory = localHistory.input
        }
        this.webview?.postMessage({
            type: 'history',
            messages: localHistory,
        })
    }
    /**
     * create webview resources
     */
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
        const root = vscode.Uri.joinPath(webviewPath, 'index.html')
        const bytes = await vscode.workspace.fs.readFile(root)
        const decoded = new TextDecoder('utf-8').decode(bytes)
        const resources = webviewView.webview.asWebviewUri(webviewPath)
        const nonce = this.getNonce()

        webviewView.webview.html = decoded
            .replaceAll('./', `${resources.toString()}/`)
            .replace('/nonce/', nonce)
            .replace('/tos-version/', this.tosVersion.toString())

        webviewView.webview.onDidReceiveMessage(message => this.onDidReceiveMessage(message, webviewView.webview))
    }

    public async transcriptForTesting(testing: TestSupport): Promise<ChatMessage[]> {
        if (!testing) {
            console.error('used ForTesting method without test support object')
            return []
        }
        return this.transcript.toChat()
    }

    public async onConfigChange(change: string, codebase: string, serverEndpoint: string): Promise<void> {
        switch (change) {
            case 'token':
            case 'endpoint': {
                const { intentDetector, codebaseContext, chatClient } = await configureExternalServices(
                    this.contextType,
                    codebase,
                    this.rgPath,
                    serverEndpoint,
                    this.secretStorage,
                    this.mode
                )

                this.intentDetector = intentDetector
                this.codebaseContext = codebaseContext
                this.chat = chatClient

                vscode.window.showInformationMessage('Cody configuration has been updated.')
                break
            }
        }
    }

    private getNonce(): string {
        let text = ''
        const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
        for (let i = 0; i < 32; i++) {
            text += possible.charAt(Math.floor(Math.random() * possible.length))
        }
        return text
    }

    private fixOpenMarkdownCodeBlock(text: string): string {
        const occurances = text.split('```').length - 1
        if (occurances % 2 === 1) {
            return text + '\n```'
        }
        return text
    }
}
