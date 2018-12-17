import * as clientType from '@sourcegraph/extension-api-types'
import { take } from 'rxjs/operators'
import { Range } from '../extension/types/range'
import { integrationTestContext } from './testHelpers'

describe('CodeEditor (integration)', () => {
    describe('setDecorations', () => {
        test('adds decorations', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Set some decorations and check they are present on the client.
            const codeEditor = extensionHost.app.windows[0].visibleViewComponents[0]
            codeEditor.setDecorations(null, [
                {
                    range: new Range(1, 2, 3, 4),
                    backgroundColor: 'red',
                },
            ])
            await extensionHost.internal.sync()
            expect(
                await services.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise()
            ).toEqual([
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    backgroundColor: 'red',
                },
            ] as clientType.TextDocumentDecoration[])

            // Clear the decorations and ensure they are removed.
            codeEditor.setDecorations(null, [])
            await extensionHost.internal.sync()
            expect(
                await services.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(null)
        })
    })
})
