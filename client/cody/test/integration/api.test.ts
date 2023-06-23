import * as assert from 'assert'

import * as vscode from 'vscode'

import { VSCodeHistory } from '../../src/autocomplete/history'

suite('API tests', () => {
    test('Cody registers some commands', async () => {
        const commands = await vscode.commands.getCommands(true)
        const codyCommands = commands.filter(command => command.includes('cody.'))
        assert.ok(codyCommands.length)
    })

    test('History', () => {
        const h = new VSCodeHistory(() => null)
        h.addItem({
            uri: vscode.Uri.file('foo.ts').toString(),
            languageId: 'ts',
        })
        h.addItem({
            uri: vscode.Uri.file('bar.ts').toString(),
            languageId: 'ts',
        })
        h.addItem({
            uri: vscode.Uri.file('foo.ts').toString(),
            languageId: 'ts',
        })
        assert.deepStrictEqual(
            h.lastN(20).map(h => vscode.Uri.parse(h.uri).fsPath),
            ['/foo.ts', '/bar.ts']
        )
    })
})
