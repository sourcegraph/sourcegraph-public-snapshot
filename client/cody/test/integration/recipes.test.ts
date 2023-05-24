import * as assert from 'assert'

import * as vscode from 'vscode'

import { afterIntegrationTest, beforeIntegrationTest, getTranscript } from './helpers'

suite('Recipes', function () {
    this.beforeEach(() => beforeIntegrationTest())
    this.afterEach(() => afterIntegrationTest())

    test('Explain Code', async () => {
        // Open Main.java
        assert.ok(vscode.workspace.workspaceFolders)
        const mainJavaUri = vscode.Uri.parse(`${vscode.workspace.workspaceFolders[0].uri.toString()}/Main.java`)
        const textEditor = await vscode.window.showTextDocument(mainJavaUri)

        // Select the "main" method
        textEditor.selection = new vscode.Selection(5, 0, 7, 0)

        // Run the "explain" command
        await vscode.commands.executeCommand('cody.recipe.explain-code-high-level')

        // Check the chat transcript contains markdown
        const humanMessage = await getTranscript(0)
        assert.match(humanMessage.displayText || '', /^Explain the following code/)
        assert.match(humanMessage.displayText || '', /public/)

        assert.match((await getTranscript(1)).displayText || '', /^hello from the assistant$/)
    })
})
