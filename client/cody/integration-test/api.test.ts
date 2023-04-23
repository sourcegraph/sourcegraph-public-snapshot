import * as assert from 'assert'

import * as vscode from 'vscode'

import { History } from '../src/completions/history'

suite('API tests', () => {
    test('Cody registers some commands', async () => {
        const commands = await vscode.commands.getCommands(true)
        const codyCommands = commands.filter(command => command.includes('cody.'))
        assert.ok(codyCommands.length)
    })

    test('History', () => {
        const h = new History(() => null)
        h.addItem({
            document: {
                uri: vscode.Uri.file('foo.ts'),
                languageId: 'ts',
            },
        })
        h.addItem({
            document: {
                uri: vscode.Uri.file('bar.ts'),
                languageId: 'ts',
            },
        })
        h.addItem({
            document: {
                uri: vscode.Uri.file('foo.ts'),
                languageId: 'ts',
            },
        })
        assert.deepStrictEqual(
            h.lastN(20).map(h => h.document.uri.fsPath),
            ['/foo.ts', '/bar.ts']
        )
    })
})
