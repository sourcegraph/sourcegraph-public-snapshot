import * as assert from 'assert'
import * as path from 'path'
import * as vscode from 'vscode'

import { ExtensionApi } from '../../extension-api'

async function enableCodyWithAccessToken(token: string) {
	const config = vscode.workspace.getConfiguration()
	await config.update('sourcegraph.cody.enable', true)
	await vscode.commands.executeCommand('cody.set-access-token', token)
}

async function setMockServerConfig() {
	const config = vscode.workspace.getConfiguration()

	// TODO: These ports are also hard-coded in runTest, move them into
	// a common module.
	const serverPort = 49300
	await config.update('cody.serverEndpoint', `http://localhost:${serverPort}`)

	const embeddingPort = 49301
	await config.update('cody.embeddingsEndpoint', `http://localhost:${embeddingPort}`)
}

suite('End-to-end', () => {
	vscode.window.showInformationMessage('Starting end-to-end tests.')

	test('Cody registers some commands', async () => {
		let commands = await vscode.commands.getCommands()
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
		await vscode.commands.executeCommand('cody.recipe.explain-code-high-level')

		// TODO: Wait for an indication there's a message in progress instead of racing

		// Check the chat transcript contains markdown
		let api = vscode.extensions.getExtension<ExtensionApi>('hpargecruos.kodj')
		let transcript = await api?.exports?.testing?.chatTranscript()
		assert.ok(transcript)
		assert.strictEqual(transcript.length, 1)
		assert.ok(transcript[0].displayText.startsWith('<p>Explain the following code'))
		assert.ok(transcript[0].displayText.indexOf('<span class="hljs-keyword">public</span>') !== -1)

		// TODO: Mock a response from Cody and check the transcript is updated
	})
})
