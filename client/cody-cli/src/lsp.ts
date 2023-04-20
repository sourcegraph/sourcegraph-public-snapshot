import { TextDocument } from 'vscode-languageserver-textdocument'
import {
    createConnection,
    TextDocuments,
    Diagnostic,
    DiagnosticSeverity,
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
} from 'vscode-languageserver/node'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { SourcegraphIntentDetectorClient } from '@sourcegraph/cody-shared/src/intent-detector/client'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'

import { DEFAULTS } from './config'
import { createCodebaseContext } from './context'

export async function startLSP() {
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
    connection: Connection
    documents: TextDocuments<TextDocument>

    globalSettings: CodyLSPSettings = defaultSettings
    documentSettings: Map<string, Thenable<CodyLSPSettings>> = new Map()

    hasDiagnosticRelatedInformationCapability: boolean = false

    // These 3 will be set once we've received the configuration from the LSP
    // client.
    intentDetector?: IntentDetector
    codebaseContext?: CodebaseContext
    completionsClient?: SourcegraphNodeCompletionsClient

    constructor() {
        this.connection = createConnection(ProposedFeatures.all)
        this.documents = new TextDocuments(TextDocument)

        this.connection.onInitialized(this.onInitialized.bind(this))
        this.connection.onDidChangeConfiguration(this.onDidChangeConfiguration.bind(this))
        this.connection.onCompletion(this.onCompletion.bind(this))
        this.connection.onDidChangeWatchedFiles(this.onDidChangeWatchedFiles.bind(this))
        this.connection.onCompletionResolve(this.onCompletionResolve.bind(this))

        this.documents.onDidClose(e => this.documentSettings.delete(e.document.uri))
        this.documents.onDidChangeContent(change => this.validateTextDocument(change.document))
    }

    public listen() {
        this.documents.listen(this.connection)
        this.connection.listen()
    }

    onInitialize(params: InitializeParams) {
        const capabilities = params.capabilities

        this.hasDiagnosticRelatedInformationCapability = !!(
            capabilities.textDocument &&
            capabilities.textDocument.publishDiagnostics &&
            capabilities.textDocument.publishDiagnostics.relatedInformation
        )
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

    onInitialized() {
        console.error('onInitialized')
        this.connection.client.register(DidChangeConfigurationNotification.type, undefined)
    }

    onDidChangeConfiguration(change: DidChangeConfigurationParams) {
        console.error('configuration change', change)

        this.globalSettings = (change.settings.codylsp || defaultSettings) as CodyLSPSettings
        if (validSettings(this.globalSettings.sourcegraph)) {
            this.initializeCody()
        }

        this.documents.all().forEach(this.validateTextDocument)
    }

    async initializeCody() {
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

        console.error('hey man we did it!!')
    }

    onDidChangeWatchedFiles(change: DidChangeWatchedFilesParams) {
        // Monitored files have change in VSCode
        console.error('We received an file change event:', change)
    }

    async validateTextDocument(textDocument: TextDocument): Promise<void> {
        const diagnostics: Diagnostic[] = [
            {
                severity: DiagnosticSeverity.Warning,
                range: {
                    start: textDocument.positionAt(0),
                    end: textDocument.positionAt(10),
                },
                message: `Cody was here`,
                source: 'codylsp',
            },
        ]

        this.connection.sendDiagnostics({ uri: textDocument.uri, diagnostics })
    }

    onCompletion(_textDocumentPosition: TextDocumentPositionParams): CompletionItem[] {
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

    onCompletionResolve(item: CompletionItem): CompletionItem {
        if (item.data === 1) {
            item.detail = 'TypeScript details'
            item.documentation = 'TypeScript documentation'
        } else if (item.data === 2) {
            item.detail = 'JavaScript details'
            item.documentation = 'JavaScript documentation'
        }
        return item
    }
}
