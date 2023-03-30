import path from 'path'

import * as vscode from 'vscode'

import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { getPreamble } from '@sourcegraph/cody-shared/src/chat/preamble'
import { getRecipe } from '@sourcegraph/cody-shared/src/chat/recipes'
import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { reformatBotMessage } from '@sourcegraph/cody-shared/src/chat/viewHelpers'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { version as packageVersion } from '../../package.json'
import { ChatHistory } from '../../webviews/utils/types'
import { LocalStorage } from '../command/LocalStorageProvider'
import { CODY_ACCESS_TOKEN_SECRET, getAccessToken, SecretStorage } from '../command/secret-storage'
import { updateConfiguration } from '../configuration'
import { VSCodeEditor } from '../editor/vscode-editor'
import { configureExternalServices } from '../external-services'
import { getRootPath } from '../keyword-context/local-keyword-context-fetcher'
import { getRgPath } from '../rg'
import { TestSupport } from '../test-support'

async function isValidLogin(serverEndpoint: string, accessToken: string): Promise<boolean> {
    const client = new SourcegraphGraphQLAPIClient(serverEndpoint, accessToken)
    const userId = await client.getCurrentUserId()
    return !isError(userId)
}

export class ChatViewProvider implements vscode.WebviewViewProvider {
    private isMessageInProgress = false
    private cancelCompletionCallback: (() => void) | null = null
    private webview?: vscode.Webview

    private tosVersion = packageVersion

    private currentChatID = ''
    private inputHistory: string[] = []
    private chatHistory: ChatHistory = {}

    constructor(
        private extensionPath: string,
        private codebase: string,
        private transcript: Transcript,
        private chat: ChatClient,
        private intentDetector: IntentDetector,
        private codebaseContext: CodebaseContext,
        private editor: Editor,
        private secretStorage: SecretStorage,
        private contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        private rgPath: string,
        private mode: 'development' | 'production',
        private localStorage: LocalStorage
    ) {
        if (TestSupport.instance) {
            TestSupport.instance.chatViewProvider.set(this)
        }
        // chat id is used to identify chat session
        this.createNewChatID()
    }

    public static async create(
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
        const editor = new VSCodeEditor()

        const { intentDetector, codebaseContext, chatClient } = await configureExternalServices(
            serverEndpoint,
            codebase,
            rgPath,
            editor,
            secretStorage,
            contextType,
            mode
        )
        return new ChatViewProvider(
            extensionPath,
            codebase,
            new Transcript(),
            chatClient,
            intentDetector,
            codebaseContext,
            editor,
            secretStorage,
            contextType,
            rgPath,
            mode,
            localStorage
        )
    }

    private async onDidReceiveMessage(message: any): Promise<void> {
        const rootPath = getRootPath()
        switch (message.command) {
            case 'initialized':
                await this.sendToken()
                await this.sendTranscript()
                await this.sendChatHistory()
                break
            case 'reset':
                await this.onResetChat()
                await this.sendChatHistory()
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
            case 'settings': {
                const isValid = await isValidLogin(message.serverEndpoint, message.accessToken)
                if (isValid) {
                    await updateConfiguration('serverEndpoint', message.serverEndpoint)
                    await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.accessToken)
                }
                await this.sendLogin(isValid)
                break
            }
            case 'removeToken':
                await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
                break
            case 'removeHistory':
                await this.localStorage.removeChatHistory()
                break
            case 'links':
                await vscode.env.openExternal(vscode.Uri.parse(message.value))
                break
            case 'openFile':
                if (rootPath !== null) {
                    const uri = vscode.Uri.file(path.join(rootPath, message.filePath))
                    // This opens the file in the active column.
                    try {
                        const doc = await vscode.workspace.openTextDocument(uri)
                        await vscode.window.showTextDocument(doc)
                    } catch (error) {
                        console.error(`Could not open file: ${error}`)
                    }
                } else {
                    console.error('Could not open file because rootPath is null')
                }
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

    private sendPrompt(promptMessages: Message[], responsePrefix = ''): void {
        this.cancelCompletion()

        this.cancelCompletionCallback = this.chat.chat(promptMessages, {
            onChange: text => this.onBotMessageChange(reformatBotMessage(text, responsePrefix)),
            onComplete: () => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                this.onBotMessageComplete()
            },
            onError: err => {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                vscode.window.showErrorMessage(err)
            },
        })
    }

    private cancelCompletion(): void {
        this.cancelCompletionCallback?.()
        this.cancelCompletionCallback = null
    }

    private async onResetChat(): Promise<void> {
        this.createNewChatID()
        this.cancelCompletion()
        this.isMessageInProgress = false
        this.transcript.reset()
        await this.sendTranscript()
    }

    private async onHumanMessageSubmitted(text: string): Promise<void> {
        this.inputHistory.push(text)
        await this.executeRecipe('chat-question', text)
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

        await this.showTab('chat')
        await this.sendTranscript()

        const prompt = await this.transcript.toPrompt(getPreamble(this.codebase))
        this.sendPrompt(prompt, interaction.getAssistantMessage().prefix ?? '')
    }

    private async onBotMessageChange(text: string): Promise<void> {
        this.transcript.addAssistantResponse(text)
        await this.sendTranscript()
    }

    private async onBotMessageComplete(): Promise<void> {
        this.isMessageInProgress = false
        this.cancelCompletionCallback = null
        await this.sendTranscript()
        await this.saveChatHistory()
    }

    private async showTab(tab: string): Promise<void> {
        await vscode.commands.executeCommand('cody.chat.focus')
        await this.webview?.postMessage({ type: 'showTab', tab })
    }

    private async sendTranscript(): Promise<void> {
        await this.webview?.postMessage({
            type: 'transcript',
            messages: this.transcript.toChat(),
            isMessageInProgress: this.isMessageInProgress,
        })
    }

    private async sendLogin(isValid: boolean): Promise<void> {
        await this.webview?.postMessage({ type: 'login', isValid })
    }
    /**
     * Sends access token to webview
     */
    private async sendToken(): Promise<void> {
        await this.webview?.postMessage({
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
    private async sendChatHistory(): Promise<void> {
        const localHistory = this.localStorage.getChatHistory()
        if (localHistory) {
            this.chatHistory = localHistory.chat
            this.inputHistory = localHistory.input
        }
        await this.webview?.postMessage({
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

        //   Create Webview
        const root = vscode.Uri.joinPath(webviewPath, 'index.html')
        const bytes = await vscode.workspace.fs.readFile(root)
        const decoded = new TextDecoder('utf-8').decode(bytes)
        const resources = webviewView.webview.asWebviewUri(webviewPath)
        const nonce = this.getNonce()

        webviewView.webview.html = decoded
            .replaceAll('./', `${resources.toString()}/`)
            .replace('/nonce/', nonce)
            .replace('/tos-version/', this.tosVersion.toString())
        webviewView.webview.onDidReceiveMessage(message => this.onDidReceiveMessage(message))
    }

    public transcriptForTesting(testing: TestSupport): ChatMessage[] {
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
                    serverEndpoint,
                    codebase,
                    this.rgPath,
                    this.editor,
                    this.secretStorage,
                    this.contextType,
                    this.mode
                )

                this.codebase = codebase
                this.intentDetector = intentDetector
                this.codebaseContext = codebaseContext
                this.chat = chatClient

                const action = await vscode.window.showInformationMessage(
                    'Cody configuration has been updated.',
                    'Reload Window'
                )
                if (action === 'Reload Window') {
                    await vscode.commands.executeCommand('workbench.action.reloadWindow')
                }
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
}
