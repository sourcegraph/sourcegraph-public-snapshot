import * as util from 'util'

import { BotResponseMultiplexer } from '@sourcegraph/cody-shared/src/chat/bot-response-multiplexer'
import { Client, createClient } from '@sourcegraph/cody-shared/src/chat/client'
import { getRecipe, registeredRecipes } from '@sourcegraph/cody-shared/src/chat/recipes/agent-recipes'
import { RecipeContext } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import {
    ActiveTextEditor,
    ActiveTextEditorSelection,
    ActiveTextEditorViewControllers,
    ActiveTextEditorVisibleContent,
    Editor,
} from '@sourcegraph/cody-shared/src/editor'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { MessageHandler, StaticEditor, StaticRecipeContext } from './rpc'

export class AgentIntentDetector implements IntentDetector {
    constructor(private agent: Agent) {}

    isCodebaseContextRequired(input: string): Promise<boolean | Error> {
        return this.agent.request('intent/isCodebaseContextRequired', input)
    }
    isEditorContextRequired(input: string): Promise<boolean | Error> {
        return this.agent.request('intent/isEditorContextRequired', input)
    }
}

export class AgentEditor implements Editor {
    // TODO
    controllers?: ActiveTextEditorViewControllers | undefined

    constructor(private agent: Agent, private staticEditor: StaticEditor) {}

    async getWorkspaceRootPath(): Promise<string | null> {
        return this.staticEditor.workspaceRoot
    }

    getActiveTextEditor(): Promise<ActiveTextEditor | null> {
        return this.agent.request('editor/active', void {})
    }

    getActiveTextEditorSelection(): Promise<ActiveTextEditorSelection | null> {
        return this.agent.request('editor/selection', void {})
    }

    getActiveTextEditorSelectionOrEntireFile(): Promise<ActiveTextEditorSelection | null> {
        return this.agent.request('editor/selectionOrEntireFile', void {})
    }

    getActiveTextEditorVisibleContent(): Promise<ActiveTextEditorVisibleContent | null> {
        return this.agent.request('editor/visibleContent', void {})
    }

    async replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void> {
        // Handle possible failure
        await this.agent.request('editor/replaceSelection', {
            fileName,
            selectedText,
            replacement,
        })
    }

    async showQuickPick(labels: string[]): Promise<string | null> {
        return await this.agent.request('editor/quickPick', labels)
    }

    async showWarningMessage(message: string): Promise<void> {
        this.agent.notify('editor/warning', message)
    }

    async showInputBox(prompt?: string | undefined): Promise<string | null> {
        return await this.agent.request('editor/prompt', prompt!)
    }
}

export class Agent extends MessageHandler {
    private client: Promise<Client>

    constructor() {
        super()

        const agent = this

        this.client = createClient({
            editor: new AgentEditor(this, {
                workspaceRoot: null,
            }),
            config: {
                customHeaders: {},
                accessToken: process.env.SRC_ACCESS_TOKEN!,
                serverEndpoint: 'https://sourcegraph.sourcegraph.com',
                useContext: 'none',
            },
            async setMessageInProgress(messageInProgress) {
                agent.notify('chat/updateMessageInProgress', messageInProgress)
            },
            async setTranscript(transcript) {
                agent.notify('chat/updateTranscript', await transcript.toJSON())
            },
            CompletionsClient: SourcegraphNodeCompletionsClient,
        })

        this.registerRequest('initialize', async client => {
            process.stderr.write(`Beginning handshake with client ${client.name}\n`)
            return {
                name: 'cody-agent',
            }
        })

        this.registerRequest('shutdown', async client => {})

        this.registerNotification('exit', async client => {
            process.exit(0)
        })

        this.registerRequest('recipes/list', async data => {
            return Object.values(registeredRecipes).map(({ id, title }) => ({
                id,
                title,
            }))
        })

        this.registerRequest('recipes/execute', async data => {
            ;(await this.client).executeRecipe(data.id, {
                humanChatInput: data.humanChatInput,
            })
        })
    }
}
