import assert from 'assert'
import { spawn } from 'child_process'
import path from 'path'

import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'

import { MessageHandler } from './jsonrpc'

export class TestClient extends MessageHandler {
    public async handshake() {
        const info = await this.request('initialize', {
            name: 'test-client',
            version: 'v1',
            workspaceRootPath: '/path/to/foo',
        })
        this.notify('initialized', null)
        return info
    }

    public listRecipes() {
        return this.request('recipes/list', null)
    }

    public async executeRecipe(id: RecipeID, humanChatInput: string) {
        return this.request('recipes/execute', {
            id,
            humanChatInput,
        })
    }

    public async shutdownAndExit() {
        await this.request('shutdown', null)
        this.notify('exit', null)
    }
}

describe('StandardAgent', () => {
    if (process.env.SRC_ACCESS_TOKEN === undefined || process.env.SRC_ENDPOINT === undefined) {
        it('no-op test because SRC_ACCESS_TOKEN is not set. To actually run the Cody Agent tests, set the environment variables SRC_ENDPOINT and SRC_ACCESS_TOKEN', () => {})
        return
    }
    const client = new TestClient()
    const agentProcess = spawn('node', [path.join(__dirname, '../dist/agent.js')], {
        stdio: 'pipe',
    })

    agentProcess.stdout.pipe(client.messageDecoder)
    client.messageEncoder.pipe(agentProcess.stdin)
    agentProcess.stderr.on('data', msg => {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call
        console.log(msg.toString())
    })

    it('initializes properly', async () => {
        assert.deepStrictEqual(await client.handshake(), { name: 'cody-agent' }, 'Agent should be cody-agent')
    })

    it('lists recipes correctly', async () => {
        const recipes = await client.listRecipes()
        assert(recipes.length === 8)
    })

    const streamingChatMessages = new Promise<void>(resolve => {
        client.registerNotification('chat/updateMessageInProgress', msg => {
            if (msg === null) {
                resolve()
            }
        })
    })

    it('allows us to execute recipes properly', async () => {
        await client.executeRecipe('chat-question', "What's 2+2?")
    })

    it('sends back transcript updates and makes sense', () => streamingChatMessages, 20_000)

    afterAll(async () => {
        await client.shutdownAndExit()
    })
})
