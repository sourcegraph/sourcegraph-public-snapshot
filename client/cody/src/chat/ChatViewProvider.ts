import path from 'path'

import * as vscode from 'vscode'

import { BotResponseMultiplexer } from '@sourcegraph/cody-shared/src/chat/bot-response-multiplexer'
import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { getPreamble } from '@sourcegraph/cody-shared/src/chat/preamble'
import { getRecipe } from '@sourcegraph/cody-shared/src/chat/recipes/vscode-recipes'
import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { ChatMessage, ChatHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { reformatBotMessage } from '@sourcegraph/cody-shared/src/chat/viewHelpers'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { highlightTokens } from '@sourcegraph/cody-shared/src/hallucinations-detector'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { View } from '../../webviews/NavBar'
import { LocalStorage } from '../command/LocalStorageProvider'
import { updateConfiguration } from '../configuration'
import { logEvent } from '../event-logger'
import { CODY_ACCESS_TOKEN_SECRET, SecretStorage } from '../secret-storage'
import { TestSupport } from '../test-support'

import { ConfigurationSubsetForWebview, DOTCOM_URL, ExtensionMessage, WebviewMessage } from './protocol'

export async function isValidLogin(
    config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
): Promise<boolean> {
    const client = new SourcegraphGraphQLAPIClient(config)
    const userId = await client.getCurrentUserId()
    return !isError(userId)
}

type Config = Pick<
    ConfigurationWithAccessToken,
    'codebase' | 'serverEndpoint' | 'debug' | 'customHeaders' | 'accessToken' | 'useContext'
>

export class ChatViewProvider implements vscode.WebviewViewProvider, vscode.Disposable {
    private isMessageInProgress = false
    private cancelCompletionCallback: (() => void) | null = null
    private webview?: Omit<vscode.Webview, 'postMessage'> & {
        postMessage(message: ExtensionMessage): Thenable<boolean>
    }

    private currentChatID = ''
    private inputHistory: string[] = []
    private chatHistory: ChatHistory = {}

    private transcript: Transcript = new Transcript()

    // Allows recipes to hook up subscribers to process sub-streams of bot output
    private multiplexer: BotResponseMultiplexer = new BotResponseMultiplexer()

    private configurationChangeEvent = new vscode.EventEmitter<void>()

    private disposables: vscode.Disposable[] = []

    constructor(
        private extensionPath: string,
        private config: Config,
        private chat: ChatClient,
        private intentDetector: IntentDetector,
        private codebaseContext: CodebaseContext,
        private editor: Editor,
        private secretStorage: SecretStorage,
        private localStorage: LocalStorage
    ) {
        if (TestSupport.instance) {
            TestSupport.instance.chatViewProvider.set(this)
        }
        // chat id is used to identify chat session
        this.createNewChatID()

        this.disposables.push(this.configurationChangeEvent)
    }

    public onConfigurationChange(newConfig: Config): void {
        this.config = newConfig
        this.configurationChangeEvent.fire()
    }

    public clearAndRestartSession(): void {
        this.createNewChatID()
        this.cancelCompletion()
        this.isMessageInProgress = false
        this.transcript.reset()
        this.sendTranscript()
        this.sendChatHistory()
    }

    private async onDidReceiveMessage(message: WebviewMessage): Promise<void> {
        switch (message.command) {
            case 'initialized':
                this.publishContextStatus()
                this.publishConfig()
                this.sendTranscript()
                this.sendChatHistory()
                break
            case 'submit':
                await this.onHumanMessageSubmitted(message.text)
                break
            case 'edit':
                this.transcript.removeLastInteraction()
                await this.onHumanMessageSubmitted(message.text)
                break
            case 'executeRecipe':
                await this.executeRecipe(message.recipe)
                break
            case 'settings': {
                const isValid = await isValidLogin({
                    serverEndpoint: message.serverEndpoint,
                    accessToken: message.accessToken,
                    customHeaders: this.config.customHeaders,
                })
                // activate when user has valid login
                await vscode.commands.executeCommand('setContext', 'cody.activated', isValid)
                if (isValid) {
                    await updateConfiguration('serverEndpoint', message.serverEndpoint)
                    await this.secretStorage.store(CODY_ACCESS_TOKEN_SECRET, message.accessToken)
                    this.sendEvent('auth', 'login')
                }
                void this.webview?.postMessage({ type: 'login', isValid })
                break
            }
            case 'event':
                this.sendEvent(message.event, message.value)
                break
            case 'removeToken':
                await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
                this.sendEvent('token', 'Delete')
                break
            case 'removeHistory':
                await this.localStorage.removeChatHistory()
                break
            case 'links':
                void this.openExternalLinks(message.value)
                break
            case 'openFile': {
                const rootPath = this.editor.getWorkspaceRootPath()
                if (!rootPath) {
                    this.sendErrorToWebview('Failed to open file: missing rootPath')
                    return
                }
                try {
                    // This opens the file in the active column.
                    const uri = vscode.Uri.file(path.join(rootPath, message.filePath))
                    const doc = await vscode.workspace.openTextDocument(uri)
                    await vscode.window.showTextDocument(doc)
                } catch {
                    // Try to open the file in the sourcegraph view
                    const sourcegraphSearchURL = new URL(
                        `/search?q=context:global+file:${message.filePath}`,
                        this.config.serverEndpoint
                    ).href
                    void this.openExternalLinks(sourcegraphSearchURL)
                }
                break
            }
            default:
                this.sendErrorToWebview('Invalid request type from Webview')
        }
    }

    private createNewChatID(): void {
        this.currentChatID = new Date(Date.now()).toUTCString()
    }

    private sendPrompt(promptMessages: Message[], responsePrefix = ''): void {
        this.cancelCompletion()

        let text = ''

        this.multiplexer.sub(BotResponseMultiplexer.DEFAULT_TOPIC, {
            onResponse: (content: string) => {
                text += content
                this.transcript.addAssistantResponse(reformatBotMessage(text, responsePrefix))
                this.sendTranscript()
                return Promise.resolve()
            },
            onTurnComplete: async () => {
                const lastInteraction = this.transcript.getLastInteraction()
                if (lastInteraction) {
                    const { text, displayText } = lastInteraction.getAssistantMessage()
                    const { text: highlightedDisplayText } = await highlightTokens(displayText || '', fileExists)
                    this.transcript.addAssistantResponse(text || '', highlightedDisplayText)
                }
                void this.onCompletionEnd()
            },
        })

        let textConsumed = 0

        this.cancelCompletionCallback = this.chat.chat(promptMessages, {
            onChange: text => {
                // TODO(dpc): The multiplexer can handle incremental text. Change chat to provide incremental text.
                text = text.slice(textConsumed)
                textConsumed += text.length
                return this.multiplexer.publish(text)
            },
            onComplete: () => this.multiplexer.notifyTurnComplete(),
            onError: (err, statusCode) => {
                // Display error message as assistant response
                this.transcript.addErrorAsAssistantResponse(
                    `<div class="cody-chat-error"><span>Request failed: </span>${err}</div>`
                )
                // Log users out on unauth error
                if (statusCode && statusCode >= 400 && statusCode <= 410) {
                    void this.sendLogin(false)
                    this.clearAndRestartSession()
                }
                this.onCompletionEnd()
                console.error(`Completion request failed: ${err}`)
            },
        })
    }

    private cancelCompletion(): void {
        this.cancelCompletionCallback?.()
        this.cancelCompletionCallback = null
    }

    private onCompletionEnd(): void {
        this.isMessageInProgress = false
        this.cancelCompletionCallback = null
        this.sendTranscript()
        void this.saveChatHistory()
    }

    private async onHumanMessageSubmitted(text: string): Promise<void> {
        this.inputHistory.push(text)
        await this.executeRecipe('chat-question', text)
    }

    public async executeRecipe(recipeId: string, humanChatInput: string = ''): Promise<void> {
        if (this.isMessageInProgress) {
            this.sendErrorToWebview('Cannot execute multiple recipes. Please wait for the current recipe to finish.')
        }

        const recipe = getRecipe(recipeId)
        if (!recipe) {
            return
        }

        // Create a new multiplexer to drop any old subscribers
        this.multiplexer = new BotResponseMultiplexer()

        const interaction = await recipe.getInteraction(humanChatInput, {
            editor: this.editor,
            intentDetector: this.intentDetector,
            codebaseContext: this.codebaseContext,
            responseMultiplexer: this.multiplexer,
        })
        if (!interaction) {
            return
        }
        this.isMessageInProgress = true
        this.transcript.addInteraction(interaction)

        this.showTab('chat')
        this.sendTranscript()

        const prompt = await this.transcript.toPrompt(getPreamble(this.config.codebase))
        this.sendPrompt(prompt, interaction.getAssistantMessage().prefix ?? '')

        logEvent(`CodyVSCodeExtension:recipe:${recipe.id}:executed`)
    }

    private showTab(tab: string): void {
        void vscode.commands.executeCommand('cody.chat.focus')
        void this.webview?.postMessage({ type: 'showTab', tab })
    }

    /**
     * Send transcript to webview
     */
    private sendTranscript(): void {
        void this.webview?.postMessage({
            type: 'transcript',
            messages: this.transcript.toChat(),
            isMessageInProgress: this.isMessageInProgress,
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
     * Save Login state to webview
     */
    public async sendLogin(isValid: boolean): Promise<void> {
        await vscode.commands.executeCommand('setContext', 'cody.activated', isValid)
        if (isValid) {
            this.sendEvent('auth', 'login')
        }
        void this.webview?.postMessage({ type: 'login', isValid })
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
        void this.webview?.postMessage({
            type: 'history',
            messages: localHistory,
        })
    }

    /**
     * Publish the current context status to the webview.
     */
    private publishContextStatus(): void {
        const send = (): void => {
            const editorContext = this.editor.getActiveTextEditor()
            void this.webview?.postMessage({
                type: 'contextStatus',
                contextStatus: {
                    mode: this.config.useContext,
                    connection: this.codebaseContext.checkEmbeddingsConnection(),
                    codebase: this.config.codebase,
                    filePath: editorContext ? vscode.workspace.asRelativePath(editorContext.filePath) : undefined,
                    supportsKeyword: true,
                },
            })
        }

        this.disposables.push(vscode.window.onDidChangeTextEditorSelection(() => send()))
        send()
    }

    /**
     * Publish the config to the webview.
     */
    private publishConfig(): void {
        const send = async (): Promise<void> => {
            // check if new configuration change is valid or not
            // log user out if config is invalid
            const isAuthed = await isValidLogin({
                serverEndpoint: this.config.serverEndpoint,
                accessToken: this.config.accessToken,
                customHeaders: this.config.customHeaders,
            })
            const configForWebview: ConfigurationSubsetForWebview = {
                debug: this.config.debug,
                serverEndpoint: this.config.serverEndpoint,
                hasAccessToken: isAuthed,
            }
            void vscode.commands.executeCommand('setContext', 'cody.activated', isAuthed)
            void this.webview?.postMessage({ type: 'config', config: configForWebview })
        }

        this.disposables.push(this.configurationChangeEvent.event(() => send()))
        send().catch(error => console.error(error))
    }

    /**
     * Log Events
     */
    public sendEvent(event: string, value: string): void {
        const isPrivateInstance = new URL(this.config.serverEndpoint).href !== DOTCOM_URL.href
        switch (event) {
            case 'feedback':
                // Only include context for dot com users with connected codebase
                logEvent(
                    `CodyVSCodeExtension:codyFeedback:${value}`,
                    null,
                    !isPrivateInstance && this.config.codebase ? this.transcript.toChat() : null
                )
                break
            case 'token':
                logEvent(`CodyVSCodeExtension:cody${value}AccessToken:clicked`)
                break
            case 'auth':
                logEvent(`CodyVSCodeExtension:${value}:clicked`)
                break
            // aditya combine this with above statemenet for auth or click
            case 'click':
                logEvent(`CodyVSCodeExtension:${value}:clicked`)
                break
        }
    }

    /**
     * Display error message in webview view as banner in chat view
     * It does not display error message as assistant response
     */
    private sendErrorToWebview(errorMsg: string): void {
        void this.webview?.postMessage({ type: 'errors', errors: errorMsg })
        console.error(errorMsg)
    }

    /**
     * Set webview view
     */
    public setWebviewView(view: View): void {
        void vscode.commands.executeCommand('cody.chat.focus')
        void this.webview?.postMessage({
            type: 'view',
            messages: view,
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

        const extensionPath = vscode.Uri.file(this.extensionPath)
        const webviewPath = vscode.Uri.joinPath(extensionPath, 'dist')

        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [webviewPath],
        }

        // Create Webview using client/cody/index.html
        const root = vscode.Uri.joinPath(webviewPath, 'index.html')
        const bytes = await vscode.workspace.fs.readFile(root)
        const decoded = new TextDecoder('utf-8').decode(bytes)
        const resources = webviewView.webview.asWebviewUri(webviewPath)

        // Set HTML for webview
        // This replace variables from the client/cody/dist/index.html with webview info
        // 1. Update URIs to load styles and scripts into webview (eg. path that starts with ./)
        // 2. Update URIs for content security policy to only allow specific scripts to be run
        webviewView.webview.html = decoded
            .replaceAll('./', `${resources.toString()}/`)
            .replaceAll('{cspSource}', webviewView.webview.cspSource)

        // Register webview
        this.disposables.push(webviewView.webview.onDidReceiveMessage(message => this.onDidReceiveMessage(message)))
    }

    /**
     * Open external links
     */
    private async openExternalLinks(uri: string): Promise<void> {
        try {
            await vscode.env.openExternal(vscode.Uri.parse(uri))
        } catch (error) {
            throw new Error(`Failed to open file: ${error}`)
        }
    }

    public transcriptForTesting(testing: TestSupport): ChatMessage[] {
        if (!testing) {
            console.error('used ForTesting method without test support object')
            return []
        }
        return this.transcript.toChat()
    }

    public dispose(): void {
        for (const disposable of this.disposables) {
            disposable.dispose()
        }
        this.disposables = []
    }
}

function trimPrefix(text: string, prefix: string): string {
    if (text.startsWith(prefix)) {
        return text.slice(prefix.length)
    }
    return text
}

function trimSuffix(text: string, suffix: string): string {
    if (text.endsWith(suffix)) {
        return text.slice(0, -suffix.length)
    }
    return text
}

async function fileExists(filePath: string): Promise<boolean> {
    const patterns = [filePath, '**/' + trimSuffix(trimPrefix(filePath, '/'), '/') + '/**']
    if (!filePath.endsWith('/')) {
        patterns.push('**/' + trimPrefix(filePath, '/') + '*')
    }
    for (const pattern of patterns) {
        const files = await vscode.workspace.findFiles(pattern, null, 1)
        if (files.length > 0) {
            return true
        }
    }
    return false
}
