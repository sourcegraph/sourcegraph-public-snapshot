import * as clientType from '@sourcegraph/extension-api-types'
import { from } from 'rxjs'
import { first, switchMap, take, toArray } from 'rxjs/operators'
import { isDefined } from '../../util/types'
import { Range } from '../extension/types/range'
import { Selection } from '../extension/types/selection'
import { assertToJSON } from '../extension/types/testHelpers'
import { integrationTestContext } from './testHelpers'

describe('CodeEditor (integration)', () => {
    describe('selection', () => {
        test('observe changes', async () => {
            const { model, extensionAPI } = await integrationTestContext()

            const setSelections = (selections: Selection[]) => {
                model.next({
                    ...model.value,
                    visibleViewComponents: [
                        {
                            type: 'CodeEditor',
                            item: { uri: 'foo', languageId: 'l1', text: 't1' },
                            selections,
                            isActive: true,
                        },
                    ],
                })
            }
            setSelections([new Selection(1, 2, 3, 4)])
            setSelections([])

            const values = await from(extensionAPI.app.windows[0].activeViewComponentChanges)
                .pipe(
                    switchMap(c => (c ? c.selectionsChanges : [])),
                    take(3),
                    toArray()
                )
                .toPromise()
            assertToJSON(values.map(v => v.map(v => Selection.fromPlain(v).toPlain())), [
                [],
                [new Selection(1, 2, 3, 4).toPlain()],
                [],
            ])
        })
    })

    describe('setDecorations', () => {
        test('adds decorations', async () => {
            const { services, extensionAPI } = await integrationTestContext()
            const dt = extensionAPI.app.createDecorationType()

            // Set some decorations and check they are present on the client.
            const activeWindow = await from(extensionAPI.app.activeWindowChanges)
                .pipe(first(isDefined))
                .toPromise()
            const codeEditor = activeWindow.visibleViewComponents[0]
            codeEditor.setDecorations(dt, [
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
            codeEditor.setDecorations(dt, [])
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

            const activeWindow = await from(extensionAPI.app.activeWindowChanges)
                .pipe(first(isDefined))
                .toPromise()
            const codeEditor = activeWindow.visibleViewComponents[0]
            codeEditor.setDecorations(dt1, [
                {
                    range: new Range(1, 2, 3, 4),
                    after: {
                        hoverMessage: 'foo',
                    },
                },
            ])
            codeEditor.setDecorations(dt2, [
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
            codeEditor.setDecorations(dt1, [
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
            codeEditor.setDecorations(dt2, [])
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
            const activeWindow = await from(extensionAPI.app.activeWindowChanges)
                .pipe(first(isDefined))
                .toPromise()
            const codeEditor = activeWindow.visibleViewComponents[0]
            codeEditor.setDecorations(dt, [
                {
                    range: new Range(1, 2, 3, 4),
                    backgroundColor: 'red',
                },
            ])

            // This call to setDecorations does not supply a type, mimicking extensions
            // that may have been developed against an older version of the API
            codeEditor.setDecorations(null as any, [
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
