import * as assert from 'assert'

import * as vscode from 'vscode'

import { ChatViewProvider } from '../../src/chat/ChatViewProvider'

import { afterIntegrationTest, beforeIntegrationTest, getExtensionAPI, getFixupTasks, getTranscript } from './helpers'

async function getChatViewProvider(): Promise<ChatViewProvider> {
    const chatViewProvider = await getExtensionAPI().exports.testing?.chatViewProvider.get()
    assert.ok(chatViewProvider)
    return chatViewProvider
}

suite('Cody Fixup Task Controller', function () {
    this.beforeEach(() => beforeIntegrationTest())
    this.afterEach(() => afterIntegrationTest())

    let textEditor = vscode.window.activeTextEditor

    // Run the non-stop recipe to create a new task before every test
    this.beforeEach(async () => {
        await vscode.commands.executeCommand('cody.chat.focus')
        const chatView = await getChatViewProvider()

        // Open index.html
        assert.ok(vscode.workspace.workspaceFolders)

        const indexUri = vscode.Uri.parse(`${vscode.workspace.workspaceFolders[0].uri.toString()}/index.html`)
        textEditor = await vscode.window.showTextDocument(indexUri)

        // Select the "title" tags to run the recipe on
        textEditor.selection = new vscode.Selection(6, 0, 7, 0)

        // Brings up the vscode input box
        await chatView.executeRecipe('non-stop', 'Replace hello with goodbye', false)

        // Check the chat transcript contains markdown
        const humanMessage = await getTranscript(0)

        assert.match(humanMessage.displayText || '', /^Cody Fixups: Replace hello with goodbye/)
        assert.match((await getTranscript(1)).displayText || '', /^Check your document for updates from Cody/)
    })

    test('task controller', async () => {
        if (textEditor === undefined) {
            assert.fail('editor is undefined')
        }
        // Check the Fixup Tasks from Task Controller contains the new task
        const tasks = await getFixupTasks()
        // Tasks length should be larger than 0
        assert.ok(tasks.length > 0)

        assert.match(tasks[0].instruction, /^Replace hello with goodbye/)

        // Get selection text from editor whch should match <title>Hello Cody</title>
        // TODO: Update to <title>Goodbye Cody</title> after we have implemented the replace method. Right now we are marking it as done
        // TODO: This needs to wait for the diff to be ready, then applied.
        const selectionText = textEditor.document.getText().trim()
        assert.match(selectionText, /.*<title>Hello Cody<\/title>.*/)

        // Run the apply command should remove all tasks from the task controller
        await vscode.commands.executeCommand('cody.fixup.apply')
        // TODO: If this really waited for apply to finish, then there would be 0 fixup tasks.
        assert.ok((await getFixupTasks()).length > 0)
    })

    test('show this fixup', async () => {
        // Check the Fixup Tasks from Task Controller contains the new task
        const tasks = await getFixupTasks()
        // Tasks length should be larger than 0
        assert.ok(tasks.length > 0)

        // Switch to a different file
        const mainJavaUri = vscode.Uri.parse(`${vscode.workspace.workspaceFolders?.[0].uri.toString()}/Main.java`)
        await vscode.workspace.openTextDocument(mainJavaUri)

        // Run show command to open fixup file with range selected
        await vscode.commands.executeCommand('cody.fixup.open', tasks[0].id)

        const newEditor = vscode.window.activeTextEditor
        assert.strictEqual(newEditor, textEditor)
    })

    test('apply fixups', async () => {
        const tasks = await getFixupTasks()
        assert.ok(tasks.length > 0)

        // Run the apply command should remove all tasks from the task controller
        await vscode.commands.executeCommand('cody.fixup.apply-all')
        // TODO: Update this test to wait for application, and then check that
        // there are no tasks. Apply all is not implemented so currently this
        // is a no-op.
        assert.ok((await getFixupTasks()).length > 0)
    })
})
