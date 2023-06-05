import assert from 'assert'
import { spawn } from 'child_process'
import path from 'path'

import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'

import { MessageHandler } from './rpc'

export class TestClient extends MessageHandler {
    constructor() {
        super()

        this.registerRequest('editor/selection', async _ => null)
        this.registerRequest('intent/isEditorContextRequired', async _ => false)
        this.registerRequest('editor/visibleContent', async _ => null)
    }

    async handshake() {
        const info = await this.request('initialize', {
            name: 'test-client',
        })
        this.notify('initialized', void {})
        return info
    }

    async listRecipes() {
        return await this.request('recipes/list', void {})
    }

    async executeRecipe(id: RecipeID, humanChatInput: string) {
        return this.request('recipes/execute', {
            id,
            context: {
                editor: {
                    workspaceRoot: null,
                },
                firstInteraction: true,
            },
            humanChatInput,
        })
    }

    async shutdownAndExit() {
        await this.request('shutdown', void {})
        this.notify('exit', void {})
    }
}

describe('StandardAgent', () => {
    const client = new TestClient()
    const agentProcess = spawn('node', [path.join(__dirname, '../dist/agent.js')], {
        stdio: 'pipe',
    })

    agentProcess.stdout.pipe(client.messageDecoder)
    client.messageEncoder.pipe(agentProcess.stdin)
    agentProcess.stderr.on('data', msg => {
        console.log(msg.toString())
    })

    it('initializes properly', async () => {
        assert.deepStrictEqual(await client.handshake(), { name: 'cody-agent' }, 'Agent should be cody-agent')
    })

    it('lists recipes correctly', async () => {
        const recipes = await client.listRecipes()
        assert(recipes.length === 8)
    })

    it('allows us to execute recipes properly', async () => {
        await client.executeRecipe('chat-question', "What's 2+2?")
    })

    it('sends back transcript updates and makes sense', () => {
        return new Promise((resolve, reject) => {
            let done = false
            let assistantMessage: string | null = null

            client.registerNotification('chat/updateMessageInProgress', async msg => {
                if (msg !== null) {
                    assistantMessage = msg.text ?? null
                } else {
                    done = true
                }
            })

            client.registerNotification('chat/updateTranscript', async transcript => {
                if (
                    done &&
                    assistantMessage ===
                        transcript.interactions[transcript.interactions.length - 1].assistantMessage.text
                ) {
                    if (assistantMessage.includes('4')) {
                        resolve(void {})
                    } else {
                        reject()
                    }
                }
            })
        })
    }, 20_000)

    afterAll(() => {
        client.shutdownAndExit()
    })
})
