import { v4 as uuidv4 } from 'uuid'
import { TextDocument } from 'vscode-languageserver-textdocument'
import {
    createConnection,
    TextDocuments,
    Diagnostic,
    // DiagnosticSeverity,
    ProposedFeatures,
    InitializeParams,
    DidChangeConfigurationNotification,
    CompletionItem,
    CompletionItemKind,
    TextDocumentPositionParams,
    TextDocumentSyncKind,
    InitializeResult,
    Connection,
    DidChangeConfigurationParams,
    DidChangeWatchedFilesParams,
    LogMessageParams,
    MessageType,
    ExecuteCommandParams,
    WorkDoneProgress,
    WorkDoneProgressBegin,
    WorkDoneProgressCreateRequest,
    ApplyWorkspaceEditRequest,
    ShowMessageNotification,
    TextDocumentIdentifier,
    TextDocumentEdit,
    TextEdit,
    Range,
    OptionalVersionedTextDocumentIdentifier,
    WorkspaceEdit,
    Position,
} from 'vscode-languageserver/node'

import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'

import { streamCompletions } from './completions'
import { DEFAULTS } from './config'
import { createCodebaseContext } from './context'
import { interactionFromMessage } from './interactions'
import { getPreamble } from './preamble'

export function startLSP() {
    const server = new CodyLanguageServer()
    server.listen()
}

interface SourcegraphSettings {
    url: string
    accessToken: string
    repos: string[]
}

function validSettings(settings: SourcegraphSettings): boolean {
    return settings.url !== '' && settings.accessToken !== '' && settings.repos.length !== 0
}

interface CodyLSPSettings {
    sourcegraph: SourcegraphSettings
}

const defaultSettings: CodyLSPSettings = {
    sourcegraph: {
        url: DEFAULTS.serverEndpoint,
        accessToken: 'not set',
        repos: [],
    },
}

class CodyLanguageServer {
    private connection: Connection
    private documents: TextDocuments<TextDocument>

    private globalSettings: CodyLSPSettings = defaultSettings
    private documentSettings: Map<string, Thenable<CodyLSPSettings>> = new Map()

    // These 3 will be set once we've received the configuration from the LSP
    // client.
    private intentDetector?: IntentDetector
    private codebaseContext?: CodebaseContext
    private completionsClient?: SourcegraphNodeCompletionsClient

    constructor() {
        this.connection = createConnection(ProposedFeatures.all)
        this.documents = new TextDocuments(TextDocument)

        this.connection.onInitialize(this.onInitialize.bind(this))
        this.connection.onInitialized(() => {
            this.onInitialized.bind(this)
        })
        this.connection.onDidChangeConfiguration(change => {
            this.onDidChangeConfiguration(change)
        })
        this.connection.onCompletion(this.onCompletion.bind(this))
        this.connection.onDidChangeWatchedFiles(this.onDidChangeWatchedFiles.bind(this))
        this.connection.onCompletionResolve(this.onCompletionResolve.bind(this))
        this.connection.onExecuteCommand(this.onExecuteCommand.bind(this))

        this.documents.onDidClose(e => this.documentSettings.delete(e.document.uri))
        this.documents.onDidChangeContent(change => this.validateTextDocument(change.document))
    }

    public listen() {
        this.documents.listen(this.connection)
        this.connection.listen()
    }

    private onInitialize(params: InitializeParams): InitializeResult {
        const result: InitializeResult = {
            capabilities: {
                textDocumentSync: TextDocumentSyncKind.Incremental,
                // Tell the client that this server supports code completion.
                completionProvider: {
                    resolveProvider: true,
                },
                inlayHintProvider: false,
            },
        }
        return result
    }

    private async onInitialized() {
        await this.connection.client.register(DidChangeConfigurationNotification.type, undefined)
    }

    private async onDidChangeConfiguration(change: DidChangeConfigurationParams) {
        this.globalSettings = (change.settings.codylsp || defaultSettings) as CodyLSPSettings
        if (validSettings(this.globalSettings.sourcegraph)) {
            await this.initializeCody()
        } else {
            this.connection.sendNotification(ShowMessageNotification.type.method, {
                message: 'Invalid settings',
                type: MessageType.Error,
            })
            return
        }

        for (const doc of this.documents.all()) {
            await this.validateTextDocument(doc)
        }
    }

    private async initializeCody() {
        // TODO: These two are clunky
        const codebase = this.globalSettings.sourcegraph.repos[0]
        const contextType = 'blended'
        const customHeaders = {}

        const sourcegraphClient = new SourcegraphGraphQLAPIClient({
            serverEndpoint: this.globalSettings.sourcegraph.url,
            accessToken: this.globalSettings.sourcegraph.accessToken,
            customHeaders,
        })

        try {
            this.codebaseContext = await createCodebaseContext(sourcegraphClient, codebase, contextType)
        } catch (error) {
            const errorMessage =
                `Cody could not connect to your Sourcegraph instance: ${error}\n` +
                'Make sure that cody.serverEndpoint is set to a running Sourcegraph instance and that an access token is configured.'

            this.connection.sendNotification(ShowMessageNotification.type.method, {
                message: errorMessage,
                type: MessageType.Error,
            })
        }

        this.intentDetector = new SourcegraphIntentDetectorClient(sourcegraphClient)

        this.completionsClient = new SourcegraphNodeCompletionsClient({
            serverEndpoint: this.globalSettings.sourcegraph.url,
            accessToken: this.globalSettings.sourcegraph.accessToken,
            debug: DEFAULTS.debug === 'development',
            customHeaders,
        })

        const params: LogMessageParams = { type: MessageType.Info, message: 'Cody LSP initialized, my friend!' }
        await this.connection.sendNotification('window/logMessage', params)
    }

    private onDidChangeWatchedFiles(change: DidChangeWatchedFilesParams) {
        // Monitored files have change in VSCode
        console.error('We received an file change event:', change)
    }

    private async validateTextDocument(textDocument: TextDocument): Promise<void> {
        const diagnostics: Diagnostic[] = [
            // {
            //     severity: DiagnosticSeverity.Warning,
            //     range: {
            //         start: textDocument.positionAt(0),
            //         end: textDocument.positionAt(10),
            //     },
            //     message: 'Cody was here',
            //     source: 'codylsp',
            // },
        ]

        await this.connection.sendDiagnostics({ uri: textDocument.uri, diagnostics })
    }

    private onCompletion(_textDocumentPosition: TextDocumentPositionParams): CompletionItem[] {
        // The pass parameter contains the position of the text document in
        // which code complete got requested. For the example we ignore this
        // info and always provide the same completion items.
        return [
            {
                label: 'TypeScript',
                kind: CompletionItemKind.Text,
                data: 1,
            },
            {
                label: 'JavaScript',
                kind: CompletionItemKind.Text,
                data: 2,
            },
        ]
    }

    private onCompletionResolve(item: CompletionItem): CompletionItem {
        if (item.data === 1) {
            item.detail = 'TypeScript details'
            item.documentation = 'TypeScript documentation'
        } else if (item.data === 2) {
            item.detail = 'JavaScript details'
            item.documentation = 'JavaScript documentation'
        }
        return item
    }

    private async onExecuteCommand(params: ExecuteCommandParams): Promise<any> {
        if (
            this.intentDetector === undefined ||
            this.codebaseContext === undefined ||
            this.completionsClient === undefined
        ) {
            this.connection.sendNotification(ShowMessageNotification.type.method, {
                message: 'Cannot execute command, because not connected to Sourcegraph instance',
                type: MessageType.Error,
            })

            return
        }

        if (params.arguments === undefined) {
            this.connection.sendNotification(ShowMessageNotification.type.method, {
                message: 'Cannot execute command, because arguments are missing',
                type: MessageType.Error,
            })
            return
        }

        // We create a unique token for this command and tell the client that we started work
        const uuid = uuidv4()
        await this.connection.sendRequest(WorkDoneProgressCreateRequest.type, { token: uuid })
        await this.connection.sendProgress(WorkDoneProgress.type, uuid, {
            kind: 'begin',
            title: params.command,
        })

        let result: any

        switch (params.command) {
            case 'cody.explain':
                result = await this.handleCodyExplain({
                    uri: params.arguments[0],
                    startLine: params.arguments[1],
                    endLine: params.arguments[2],
                    question: params.arguments[3],
                })
                break

            case 'cody.replace':
                result = await this.handleCodyReplace({
                    uri: params.arguments[0],
                    startLine: params.arguments[1],
                    endLine: params.arguments[2],
                    request: params.arguments[3],
                })
                break
        }

        // Now we tell the client that the work is done
        await this.connection.sendProgress(WorkDoneProgress.type, uuid, { kind: 'end' })

        return result
    }

    async handleCodyExplain({
        uri,
        startLine,
        endLine,
        question,
    }: {
        uri: string
        startLine: number | undefined
        endLine: number | undefined
        question: string
    }): Promise<{ response: string }> {
        const doc = this.documents.get(uri)

        let initialMessage = ''
        if (doc !== undefined && startLine !== undefined && endLine !== undefined) {
            const filetype = doc.languageId
            const snippet = doc?.getText({
                start: { line: startLine, character: 0 },
                end: { line: endLine, character: 0 },
            })

            initialMessage = `I am looking at the following code snippet:\n\n\`\`\`${filetype}\n${snippet}\n\`\`\`\n\n${question}`
        } else {
            initialMessage = `${question}`
        }

        const text = await this.getCompletion(initialMessage)
        return {
            response: text,
        }
    }

    async handleCodyReplace({
        uri,
        startLine,
        endLine,
        request,
    }: {
        uri: string
        startLine: number
        endLine: number
        request: string
    }): Promise<{ response: string } | { error: string }> {
        const doc = this.documents.get(uri)
        if (doc === undefined) {
            return { error: `cannot find the doc: ${uri}` }
        }
        const filetype = doc.languageId
        const snippet = doc?.getText({
            start: { line: startLine, character: 0 },
            end: { line: endLine, character: 0 },
        })

        const initialMessage = `I am looking at the following code snippet:\n\n\`\`\`${filetype}\n${snippet}\n\`\`\`\n\n${request}`

        const text = await this.getCompletion(initialMessage)

        // TODO: This is incredibly hacky
        const regex = /```\w+([\s\S]*?)```/g
        const match = regex.exec(text)
        const extractedStr = match && match[1] ? match[1] : ''

        const startPosition = Position.create(startLine, 0)
        const endPosition = Position.create(endLine, 9999999) // this is fucking ugly, jesus
        const textEdit = TextEdit.replace(Range.create(startPosition, endPosition), extractedStr)
        const edit = TextDocumentEdit.create({ uri: doc.uri, version: null }, [textEdit])

        await this.connection.sendRequest(ApplyWorkspaceEditRequest.type.method, {
            edit: {
                documentChanges: [edit],
            },
        })

        return {
            response: 'done',
        }
    }

    private async getCompletion(initialMessageText: string): Promise<string> {
        if (
            this.intentDetector === undefined ||
            this.codebaseContext === undefined ||
            this.completionsClient === undefined
        ) {
            this.connection.sendNotification(ShowMessageNotification.type.method, {
                message: 'cannot execute command, because not connected to Sourcegraph instance',
                type: MessageType.Error,
            })

            return ''
        }
        const transcript = new Transcript()
        const initialMessage: Message = { speaker: 'human', text: initialMessageText }
        const messages: { human: Message; assistant?: Message }[] = [{ human: initialMessage }]
        for (const [index, message] of messages.entries()) {
            const interaction = await interactionFromMessage(
                message.human,
                this.intentDetector,
                // Fetch codebase context only for the last message
                index === messages.length - 1 ? this.codebaseContext : null
            )

            transcript.addInteraction(interaction)

            if (message.assistant?.text) {
                transcript.addAssistantResponse(message.assistant?.text)
            }
        }

        const finalPrompt = await transcript.toPrompt(getPreamble(this.globalSettings.sourcegraph.repos[0]))

        let text = ''
        const completionsClient = this.completionsClient

        await new Promise<void>(resolve => {
            streamCompletions(completionsClient, finalPrompt, {
                onChange: chunk => {
                    text = chunk
                },
                onComplete: () => {
                    resolve()
                },
                onError: (err: string) => {
                    text = err
                    resolve()
                },
            })
        })

        return text
    }
}
