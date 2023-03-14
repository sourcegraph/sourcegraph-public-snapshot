import * as assert from 'assert'

import * as vscode from 'vscode'

import { ChatMessage } from '../../chat/ChatViewProvider'
import { ExtensionApi } from '../../extension-api'
import * as mockServer from '../mock-server'

async function enableCodyWithAccessToken(token: string): Promise<void> {
    const config = vscode.workspace.getConfiguration()
    await config.update('sourcegraph.cody.enable', true)
    await vscode.commands.executeCommand('cody.set-access-token', token)
}

async function setMockServerConfig(): Promise<void> {
    const config = vscode.workspace.getConfiguration()
    await config.update('cody.serverEndpoint', `http://localhost:${mockServer.SERVER_PORT}`)
    await config.update('cody.embeddingsEndpoint', `http://localhost:${mockServer.EMBEDDING_PORT}`)
}

async function waitUntil(predicate: () => Promise<boolean>): Promise<void> {
    let delay = 10
    while (!(await predicate())) {
        await new Promise(resolve => setTimeout(resolve, delay))
        delay <<= 1
    }
}

// executeCommand specifies ...any[] https://code.visualstudio.com/api/references/vscode-api#commands
// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function ensureExecuteCommand<T>(command: string, ...args: any[]): Promise<T> {
    await waitUntil(async () => (await vscode.commands.getCommands(true)).includes(command))
    const result = await vscode.commands.executeCommand<T>(command, ...args)
    return result
}

// Waits for the index-th message to appear in the chat transcript, and returns it.
async function getTranscript(api: vscode.Extension<ExtensionApi>, index: number): Promise<ChatMessage> {
    let transcript
    await waitUntil(async () => {
        transcript = await api.exports.testing?.chatTranscript()
        return Boolean(transcript && transcript.length > index)
    })
    assert.ok(transcript)
    return transcript[index]
}

suite('End-to-end', () => {
    void vscode.window.showInformationMessage('Starting end-to-end tests.')

    test('Cody registers some commands', async () => {
        const commands = await vscode.commands.getCommands(true)
        const codyCommands = commands.filter(command => command.includes('cody.'))
        assert.ok(codyCommands.length)
    })

    test('Explain Code', async () => {
        await enableCodyWithAccessToken('test-token')
        await setMockServerConfig()

        // Open Main.java
        assert.ok(vscode.workspace.workspaceFolders)
        const mainJavaUri = vscode.Uri.parse(`${vscode.workspace.workspaceFolders[0].uri.toString()}/Main.java`)
        const textEditor = await vscode.window.showTextDocument(mainJavaUri)

        // Select the "main" method
        textEditor.selection = new vscode.Selection(5, 0, 7, 0)

        // Run the "explain" command
        await ensureExecuteCommand('cody.recipe.explain-code-high-level')
        const api = vscode.extensions.getExtension<ExtensionApi>('hpargecruos.kodj')
        assert.ok(api)

        // Check the chat transcript contains markdown
        let message = await getTranscript(api, 0)
        assert.ok(message.displayText.startsWith('<p>Explain the following code'))
        assert.ok(message.displayText.includes('<span class="hljs-keyword">public</span>'))

        // Check the server response was handled
        // "hello world" is a canned response the WebSocket server
        // in runTest.js responds to all messages with
        message = await getTranscript(api, 1)
        assert.ok(message.displayText.includes('hello, world'))
    })
})
