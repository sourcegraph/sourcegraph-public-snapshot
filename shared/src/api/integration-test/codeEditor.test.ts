import { Range, Selection } from '@sourcegraph/extension-api-classes'
import * as clientType from '@sourcegraph/extension-api-types'
import { from } from 'rxjs'
import { distinctUntilChanged, first, switchMap, take, toArray } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../util/types'
import { assertToJSON, integrationTestContext } from './testHelpers'

describe('CodeEditor (integration)', () => {
    describe('selection', () => {
        test('observe changes', async () => {
            const {
                services: { editor: editorService },
                extensionAPI,
            } = await integrationTestContext()
            const editor = editorService.editors.get('editor#0')!
            editorService.setSelections(editor, [new Selection(1, 2, 3, 4)])
            editorService.setSelections(editor, [])

            const values = await from(extensionAPI.app.windows[0].activeViewComponentChanges)
                .pipe(
                    switchMap(c => (c ? c.selectionsChanges : [])),
                    distinctUntilChanged(),
                    take(3),
                    toArray()
                )
                .toPromise()
            assertToJSON(
                values.map(v => v.map(v => Selection.fromPlain(v).toPlain())),
                [[], [new Selection(1, 2, 3, 4).toPlain()], []]
            )
        })
    })

    describe('setDecorations', () => {
        test('adds decorations', async () => {
            const { services, extensionAPI } = await integrationTestContext()
            const dt = extensionAPI.app.createDecorationType()

            // Set some decorations and check they are present on the client.
            const editor = await getFirstCodeEditor(extensionAPI)
            editor.setDecorations(dt, [
                {
                    range: new Range(1, 2, 3, 4),
                    backgroundColor: 'red',
                },
            ])
            await extensionAPI.internal.sync()
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
            editor.setDecorations(dt, [])
            await extensionAPI.internal.sync()
            expect(
                await services.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(null)
        })

        it('merges decorations from several types', async () => {
            const { services, extensionAPI } = await integrationTestContext()
            const [dt1, dt2] = [extensionAPI.app.createDecorationType(), extensionAPI.app.createDecorationType()]

            const editor = await getFirstCodeEditor(extensionAPI)
            editor.setDecorations(dt1, [
                {
                    range: new Range(1, 2, 3, 4),
                    after: {
                        hoverMessage: 'foo',
                    },
                },
            ])
            editor.setDecorations(dt2, [
                {
                    range: new Range(1, 2, 3, 4),
                    after: {
                        hoverMessage: 'bar',
                    },
                },
            ])
            await extensionAPI.internal.sync()
            expect(
                await services.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise()
            ).toEqual([
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    after: {
                        hoverMessage: 'foo',
                    },
                },
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    after: {
                        hoverMessage: 'bar',
                    },
                },
            ] as clientType.TextDocumentDecoration[])

            // Change decorations only for dt1, and check that merged decorations are coherent
            editor.setDecorations(dt1, [
                {
                    range: new Range(1, 2, 3, 4),
                    after: {
                        hoverMessage: 'baz',
                    },
                },
            ])
            await extensionAPI.internal.sync()
            expect(
                await services.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise()
            ).toEqual([
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    after: {
                        hoverMessage: 'baz',
                    },
                },
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    after: {
                        hoverMessage: 'bar',
                    },
                },
            ] as clientType.TextDocumentDecoration[])

            // remove decorations for dt2, and verify that decorations for dt1 are still present
            editor.setDecorations(dt2, [])
            await extensionAPI.internal.sync()
            expect(
                await services.textDocumentDecoration
                    .getDecorations({ uri: 'file:///f' })
                    .pipe(take(1))
                    .toPromise()
            ).toEqual([
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    after: {
                        hoverMessage: 'baz',
                    },
                },
            ] as clientType.TextDocumentDecoration[])
        })

        it('is backwards compatible with extensions that do not provide a decoration type', async () => {
            const { services, extensionAPI } = await integrationTestContext()
            const dt = extensionAPI.app.createDecorationType()

            // Set some decorations and check they are present on the client.
            const editor = await getFirstCodeEditor(extensionAPI)
            editor.setDecorations(dt, [
                {
                    range: new Range(1, 2, 3, 4),
                    backgroundColor: 'red',
                },
            ])

            // This call to setDecorations does not supply a type, mimicking extensions
            // that may have been developed against an older version of the API
            editor.setDecorations(null as any, [
                {
                    range: new Range(1, 2, 3, 4),
                    after: {
                        hoverMessage: 'foo',
                    },
                },
            ])

            // Both sets of decorations should be displayed
            await extensionAPI.internal.sync()
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
                {
                    range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                    after: {
                        hoverMessage: 'foo',
                    },
                },
            ] as clientType.TextDocumentDecoration[])
        })
    })
})

async function getFirstCodeEditor(extensionAPI: typeof sourcegraph): Promise<sourcegraph.CodeEditor> {
    return from(extensionAPI.app.activeWindowChanges)
        .pipe(
            first(isDefined),
            switchMap(win => win.activeViewComponentChanges),
            first(isDefined)
        )
        .toPromise()
}
