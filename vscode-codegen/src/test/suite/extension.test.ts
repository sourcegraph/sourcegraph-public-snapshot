import * as assert from 'assert'
import * as path from 'path'
import * as vscode from 'vscode'

import * as mockServer from '../mock-server'

import { ChatMessage } from '../../chat/view'
import { ExtensionApi } from '../../extension-api'

async function enableCodyWithAccessToken(token: string) {
	const config = vscode.workspace.getConfiguration()
	await config.update('sourcegraph.cody.enable', true)
	await vscode.commands.executeCommand('cody.set-access-token', token)
}

async function setMockServerConfig() {
	const config = vscode.workspace.getConfiguration()
	await config.update('cody.serverEndpoint', `http://localhost:${mockServer.SERVER_PORT}`)
	await config.update('cody.embeddingsEndpoint', `http://localhost:${mockServer.EMBEDDING_PORT}`)
}

async function waitUntil(predicate: () => Promise<boolean>) {
	let delay = 10
	while (!(await predicate())) {
		await new Promise(resolve => setTimeout(resolve, delay))
		delay <<= 1
	}
}

async function ensureExecuteCommand(command: string, ...args: any[]) {
	waitUntil(async () => (await vscode.commands.getCommands(true)).indexOf(command) === -1)
	return vscode.commands.executeCommand(command, ...args)
}

// Waits for the i-th message to appear in the chat transcript, and returns it.
async function getTranscript(api: vscode.Extension<ExtensionApi>, i: number): Promise<ChatMessage> {
	let transcript = undefined
	await waitUntil(async function () {
		transcript = await api.exports.testing?.chatTranscript()
		return Boolean(transcript && transcript.length > i)
	})
	assert.ok(transcript)
	return transcript[i]
}

suite('End-to-end', () => {
	vscode.window.showInformationMessage('Starting end-to-end tests.')

	test('Cody registers some commands', async () => {
		let commands = await vscode.commands.getCommands(true)
		let codyCommands = commands.filter(command => command.indexOf('cody.') != -1)
		assert.ok(codyCommands.length)
	})

	test('Explain Code', async () => {
		await enableCodyWithAccessToken('test-token')
		await setMockServerConfig()

		// Open Main.java
		assert.ok(vscode.workspace.workspaceFolders)
		const mainJavaUri = vscode.Uri.parse(`${vscode.workspace.workspaceFolders[0].uri}/Main.java`)
		let textEditor = await vscode.window.showTextDocument(mainJavaUri)

		// Select the "main" method
		textEditor.selection = new vscode.Selection(5, 0, 7, 0)

		// Run the "explain" command
		await ensureExecuteCommand('cody.recipe.explain-code-high-level')
		let api = vscode.extensions.getExtension<ExtensionApi>('hpargecruos.kodj')
		assert.ok(api)

		// Check the chat transcript contains markdown
		let message = await getTranscript(api, 0)
		assert.ok(message.displayText.startsWith('<p>Explain the following code'))
		assert.ok(message.displayText.indexOf('<span class="hljs-keyword">public</span>') !== -1)

		// Check the server response was handled
		// "hello world" is a canned response the WebSocket server
		// in runTest.js responds to all messages with
		message = await getTranscript(api, 1)
		assert.ok(message.displayText.indexOf('hello, world') !== -1)
	})
})
