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
        console.error('onInitialized')
        await this.connection.client.register(DidChangeConfigurationNotification.type, undefined)
    }

    private async onDidChangeConfiguration(change: DidChangeConfigurationParams) {
        console.error('configuration change', change)

        this.globalSettings = (change.settings.codylsp || defaultSettings) as CodyLSPSettings
        if (validSettings(this.globalSettings.sourcegraph)) {
            await this.initializeCody()
        } else {
            console.error('invalid settings')
        }

        for (const doc of this.documents.all()) {
            await this.validateTextDocument(doc)
        }
    }

    private async initializeCody() {
        console.error('initializecody. globalSettings', this.globalSettings)
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
            // TODO: return this to the LSP

            let errorMessage = ''
            // if (isRepoNotFoundError(error)) {
            //     errorMessage =
            //         `Cody could not find the '${codebase}' repository on your Sourcegraph instance.\n` +
            //         'Please check that the repository exists and is entered correctly in the cody.codebase setting.'
            // } else {
            errorMessage =
                `Cody could not connect to your Sourcegraph instance: ${error}\n` +
                'Make sure that cody.serverEndpoint is set to a running Sourcegraph instance and that an access token is configured.'
            // }
            console.error(errorMessage)
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
            console.error('cannot execute command. not initialized', params.command)
            return
        }

        if (params.arguments === undefined) {
            console.error('no arguments to command given')
            return
        }

        const uri: string = params.arguments[0] as string
        const startLine: number | undefined = params.arguments[1] as number
        const endLine: number | undefined = params.arguments[2] as number
        const question: string = params.arguments[3] as string

        const doc = this.documents.get(uri)
        const snippet = doc?.getText({ start: { line: startLine, character: 0 }, end: { line: endLine, character: 0 } })

        const uuid = 'uuid-foobar'

        const param: WorkDoneProgressBegin = {
            kind: 'begin',
            title: params.command,
        }

        await this.connection.sendProgress(WorkDoneProgress.type, uuid, param)

        const initialMessageText = `I have the following code snippet:\n\n\`\`\`typescript\n${snippet}\n\`\`\`\n\n${question}`
        const text = await this.getCompletion(initialMessageText)

        await this.connection.sendProgress(WorkDoneProgress.type, uuid, { kind: 'end' })
        return {
            message: [text],
        }
    }

    private async getCompletion(initialMessageText: string): Promise<string> {
        if (
            this.intentDetector === undefined ||
            this.codebaseContext === undefined ||
            this.completionsClient === undefined
        ) {
            console.error('cannot execute command. not initialized')
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
