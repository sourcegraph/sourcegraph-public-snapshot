/* eslint-disable no-void */
import { Client, ClientInitConfig, createClient } from '@sourcegraph/cody-shared/src/chat/client'
import { registeredRecipes } from '@sourcegraph/cody-shared/src/chat/recipes/agent-recipes'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { AgentEditor } from './editor'
import { TextDocument } from './protocol'
import { MessageHandler } from './rpc'

export class Agent extends MessageHandler {
    private client: Promise<Client>
    public workspaceRootFilePath: string | null = null
    public activeDocumentFilePath: string | null = null
    public documents: Map<string, TextDocument> = new Map()

    constructor() {
        super()

        const config: ClientInitConfig = {
            customHeaders: {},
            accessToken: process.env.SRC_ACCESS_TOKEN!,
            serverEndpoint: process.env.SRC_ENDPOINT || 'https://sourcegraph.sourcegraph.com',
            useContext: 'none',
        }
        this.client = createClient({
            editor: new AgentEditor(this),
            config,
            setMessageInProgress: messageInProgress => {
                this.notify('chat/updateMessageInProgress', messageInProgress)
            },
            setTranscript: transcript => {
                transcript.toJSON().then(
                    value => this.notify('chat/updateTranscript', value),
                    () => {}
                )
            },
            createCompletionsClient: config => new SourcegraphNodeCompletionsClient(config),
        })

        this.registerRequest('initialize', client => {
            process.stderr.write(`Beginning handshake with client ${client.name}\n`)
            return Promise.resolve({
                name: 'cody-agent',
            })
        })
        this.registerNotification('initialized', () => {})

        this.registerRequest('shutdown', async () => {})

        this.registerNotification('exit', () => {
            process.exit(0)
        })

        this.registerNotification('workspaceRootPath/didChange', path => {
            this.workspaceRootFilePath = path
        })

        this.registerNotification('textDocument/didFocus', document => {
            this.activeDocumentFilePath = document.filePath
        })
        this.registerNotification('textDocument/didOpen', document => {
            this.documents.set(document.filePath, document)
            this.activeDocumentFilePath = document.filePath
        })
        this.registerNotification('textDocument/didChange', document => {
            this.documents.set(document.filePath, document)
            this.activeDocumentFilePath = document.filePath
        })
        this.registerNotification('textDocument/didClose', document => {
            this.documents.delete(document.filePath)
        })

        this.registerRequest('recipes/list', data =>
            Promise.resolve(
                Object.values(registeredRecipes).map(({ id }) => ({
                    id,
                    title: id, // TODO: will be added in a follow PR
                }))
            )
        )

        this.registerRequest('recipes/execute', async data => {
            const client = await this.client
            await client.executeRecipe(data.id, {
                humanChatInput: data.humanChatInput,
            })
        })
    }
}
