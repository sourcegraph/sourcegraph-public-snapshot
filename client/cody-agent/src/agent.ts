import { CurrentDocumentContextWithLanguage } from '@sourcegraph/cody-shared/src/autocomplete'
import { Client, createClient } from '@sourcegraph/cody-shared/src/chat/client'
import { registeredRecipes } from '@sourcegraph/cody-shared/src/chat/recipes/agent-recipes'
import { SourcegraphCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { ManualCompletionServiceAgent } from './completion/manual'
import { AgentEditor } from './editor'
import { AgentHistory } from './history'
import { MessageHandler } from './jsonrpc'
import { ConnectionConfiguration, Position, TextDocument } from './protocol'

export class Agent extends MessageHandler {
    private defaultConnectionConfig: ConnectionConfiguration = {
        customHeaders: {},
        accessToken: process.env.SRC_ACCESS_TOKEN || '',
        serverEndpoint: process.env.SRC_ENDPOINT || 'https://sourcegraph.com',
    }

    private editor: AgentEditor
    private completionsClient: SourcegraphCompletionsClient
    private client: Promise<Client>
    private manualCompletionsService: Promise<ManualCompletionServiceAgent>

    public workspaceRootPath: string | null = null
    public activeDocumentFilePath: string | null = null
    public documents: Map<string, TextDocument> = new Map()

    constructor() {
        super()

        this.editor = new AgentEditor(this)

        this.completionsClient = new SourcegraphNodeCompletionsClient({
            ...this.defaultConnectionConfig,
            debugEnable: false,
        })

        this.client = createClient({
            editor: this.editor,
            config: { ...this.defaultConnectionConfig, useContext: 'none' },
            setMessageInProgress: messageInProgress => {
                this.notify('chat/updateMessageInProgress', messageInProgress)
            },
            setTranscript: () => {
                // Not supported yet by agent.
            },
            completionsClient: this.completionsClient,
        })

        this.manualCompletionsService = new Promise(resolve => {
            this.client?.then(client => {
                new ManualCompletionServiceAgent(
                    this.editor,
                    this.completionsClient,
                    new AgentHistory(),
                    client.codebaseContext
                )
            })
        })

        this.registerRequest('initialize', async client => {
            process.stderr.write(
                `Cody Agent: handshake with client '${client.name}' (version '${client.version}') at workspace root path '${client.workspaceRootPath}'\n`
            )
            this.workspaceRootPath = client.workspaceRootPath
            if (client.connectionConfiguration) {
                ;(await this.client).onConfigurationChange({ useContext: 'none', ...client.connectionConfiguration })
            }
            return {
                name: 'cody-agent',
            }
        })
        this.registerNotification('initialized', () => {})

        this.registerRequest('shutdown', () => Promise.resolve(null))

        this.registerNotification('exit', () => {
            process.exit(0)
        })

        this.registerNotification('textDocument/didFocus', document => {
            this.activeDocumentFilePath = document.filePath
        })
        this.registerNotification('textDocument/didOpen', document => {
            this.documents.set(document.filePath, document)
            this.activeDocumentFilePath = document.filePath
        })
        this.registerNotification('textDocument/didChange', document => {
            if (document.content === undefined) {
                document.content = this.documents.get(document.filePath)?.content
            }
            this.documents.set(document.filePath, document)
            this.activeDocumentFilePath = document.filePath
        })
        this.registerNotification('textDocument/didClose', document => {
            this.documents.delete(document.filePath)
        })

        this.registerNotification('connectionConfiguration/didChange', async config => {
            ;(await this.client).onConfigurationChange({ useContext: 'none', ...config })
        })

        this.registerRequest('recipes/list', () =>
            Promise.resolve(
                Object.values(registeredRecipes).map(({ id }) => ({
                    id,
                    title: id, // TODO: will be added in a follow PR
                }))
            )
        )

        this.registerRequest('recipes/execute', async data => {
            const client = await this.client
            if (!client) {
                return null
            }
            await client.executeRecipe(data.id, {
                humanChatInput: data.humanChatInput,
            })
            return null
        })

        this.registerRequest('completions/manual', async data => {
            const man = await this.manualCompletionsService

            const ctx = this.getCurrentDocContext(
                man.tokToChar(man.maxPrefixTokens),
                man.tokToChar(man.maxSuffixTokens)
            )

            if (!ctx) {
                return null
            }

            const provider = await man.getManualCompletionProvider(ctx)

            if (!provider) {
                return null
            }

            return provider.generateCompletions(new AbortController().signal, data.count)
        })
    }

    private getCurrentDocContext(
        maxPrefixLength: number,
        maxSuffixLength: number
    ): CurrentDocumentContextWithLanguage | null {
        if (!this.activeDocumentFilePath) {
            return null
        }

        const doc = this.documents.get(this.activeDocumentFilePath)

        if (!doc || !doc.selection || !doc.content) {
            return null
        }

        const lines = doc.content.split('\n')
        const offset = positionToOffset(lines, doc.selection.start)

        let prevNonEmptyLine = ''
        for (let line = doc.selection.start.line - 1; line >= 0; line--) {
            if (lines[line].trim().length !== 0) {
                prevNonEmptyLine = lines[line]
                break
            }
        }

        let nextNonEmptyLine = ''
        for (let line = doc.selection.start.line + 1; line < lines.length; line++) {
            if (lines[line].trim().length !== 0) {
                nextNonEmptyLine = lines[line]
                break
            }
        }

        return {
            languageId: 'TODO',
            markdownLanguage: 'TODO',
            prefix: doc.content.slice(Math.max(0, offset - maxPrefixLength), offset),
            suffix: doc.content.slice(offset, offset + maxSuffixLength),
            prevLine: doc.selection.start.line === 0 ? '' : lines[doc.selection.start.line - 1],
            prevNonEmptyLine,
            nextNonEmptyLine,
        }
    }
}

function positionToOffset(lines: string[], position: Position) {
    let offset = 0
    for (let i = 0; i < position.line; i++) {
        offset += lines[i].length + 1
    }
    offset += position.character
    return offset
}
