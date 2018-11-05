import * as assert from 'assert'
import { take } from 'rxjs/operators'
import { Range } from '../extension/types/range'
import * as plain from '../protocol/plainTypes'
import { integrationTestContext } from './helpers.test'

describe('CodeEditor (integration)', () => {
    describe('setDecorations', () => {
        it('adds decorations', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()

            // Set some decorations and check they are present on the client.
            await ready
            const codeEditor = extensionHost.app.windows[0].visibleViewComponents[0]
            codeEditor.setDecorations(null, [
                {
                    range: new Range(1, 2, 3, 4),
                    backgroundColor: 'red',
                },
            ])
            await extensionHost.internal.sync()
            assert.deepStrictEqual(
                await clientController.registries.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise(),
                [
                    {
                        range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                        backgroundColor: 'red',
                    },
                ] as plain.TextDocumentDecoration[]
            )

            // Clear the decorations and ensure they are removed.
            codeEditor.setDecorations(null, [])
            await extensionHost.internal.sync()
            assert.deepStrictEqual(
                await clientController.registries.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise(),
                null
            )
        })
    })
})
