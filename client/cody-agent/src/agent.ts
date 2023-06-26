import { Client, createClient } from '@sourcegraph/cody-shared/src/chat/client'
import { registeredRecipes } from '@sourcegraph/cody-shared/src/chat/recipes/agent-recipes'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { AgentEditor } from './editor'
import { MessageHandler } from './jsonrpc'
import { ConnectionConfiguration, TextDocument } from './protocol'

export class Agent extends MessageHandler {
    private client?: Promise<Client>
    public workspaceRootPath: string | null = null
    public activeDocumentFilePath: string | null = null
    public documents: Map<string, TextDocument> = new Map()

    constructor() {
        super()

        this.setClient({
            customHeaders: {},
            accessToken: process.env.SRC_ACCESS_TOKEN || '',
            serverEndpoint: process.env.SRC_ENDPOINT || 'https://sourcegraph.com',
        })

        this.registerRequest('initialize', client => {
            process.stderr.write(
                `Cody Agent: handshake with client '${client.name}' (version '${client.version}') at workspace root path '${client.workspaceRootPath}'\n`
            )
            this.workspaceRootPath = client.workspaceRootPath
            if (client.connectionConfiguration) {
                this.setClient(client.connectionConfiguration)
            }
            return Promise.resolve({
                name: 'cody-agent',
            })
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

        this.registerNotification('connectionConfiguration/didChange', config => {
            this.setClient(config)
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

        this.registerRequest('completions/executeManual', data => {
            return Promise.resolve(null)
        })
    }

    private setClient(config: ConnectionConfiguration): void {
        this.client = createClient({
            editor: new AgentEditor(this),
            config: { ...config, useContext: 'none' },
            setMessageInProgress: messageInProgress => {
                this.notify('chat/updateMessageInProgress', messageInProgress)
            },
            setTranscript: () => {
                // Not supported yet by agent.
            },
            createCompletionsClient: config => new SourcegraphNodeCompletionsClient(config),
        })
    }
}
