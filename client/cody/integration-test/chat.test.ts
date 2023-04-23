import * as assert from 'assert'

import * as vscode from 'vscode'

import { ChatViewProvider } from '../src/chat/ChatViewProvider'

import { afterIntegrationTest, beforeIntegrationTest, getExtensionAPI, getTranscript } from './helpers'

async function getChatViewProvider(): Promise<ChatViewProvider> {
    const chatViewProvider = await getExtensionAPI().exports.testing?.chatViewProvider.get()
    assert.ok(chatViewProvider)
    return chatViewProvider
}

suite('Chat', function () {
    this.beforeEach(() => beforeIntegrationTest())
    this.afterEach(() => afterIntegrationTest())

    test('sends and receives a message', async () => {
        await vscode.commands.executeCommand('cody.chat.focus')
        const chatView = await getChatViewProvider()
        await chatView.executeRecipe('chat-question', 'hello from the human')

        assert.match((await getTranscript(0)).displayText || '', /^hello from the human$/)
        assert.match((await getTranscript(1)).displayText || '', /^hello from the assistant$/)
    })
})
