import { Selection } from '@sourcegraph/extension-api-types'
import { from, of, Subscribable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import { CodeEditor, createEditorService, EditorService, getActiveCodeEditorPosition } from './editorService'
import { TextModel } from './modelService'

export function createTestEditorService(
    editors: Subscribable<readonly CodeEditor[]> = of([])
): Pick<EditorService, 'editors'> {
    return { editors }
}

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('EditorService', () => {
    describe('editors', () => {
        test('merges in model data', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const editorService = createEditorService({
                    models: cold<TextModel[]>('a', { a: [{ uri: 'u', text: 't', languageId: 'l' }] }),
                })
                editorService.addEditor({ type: 'CodeEditor', resource: 'u', selections: [], isActive: true })
                expectObservable(
                    from(editorService.editors).pipe(
                        map(editors => editors.map(e => ({ resource: e.resource, model: e.model })))
                    )
                ).toBe('a', {
                    a: [
                        {
                            resource: 'u',
                            model: { uri: 'u', text: 't', languageId: 'l' },
                        },
                    ],
                })
            })
        })
    })

    test('addEditor', async () => {
        const editorService = createEditorService({ models: of([{ uri: 'u', text: 't', languageId: 'l' }]) })
        const editor = editorService.addEditor({ type: 'CodeEditor', resource: 'u', selections: [], isActive: true })
        expect(editor.editorId).toEqual('editor#0')
        expect(
            await from(editorService.editors)
                .pipe(first())
                .toPromise()
        ).toEqual([
            {
                type: 'CodeEditor',
                editorId: 'editor#0',
                resource: 'u',
                selections: [],
                isActive: true,
                model: { uri: 'u', text: 't', languageId: 'l' },
            },
        ])
    })

    test('setSelections', async () => {
        const editorService = createEditorService({ models: of([{ uri: 'u', text: 't', languageId: 'l' }]) })
        const editor = editorService.addEditor({ type: 'CodeEditor', resource: 'u', selections: [], isActive: true })
        const SELECTIONS: Selection[] = [
            {
                start: { line: 3, character: -1 },
                end: { line: 3, character: -1 },
                anchor: { line: 3, character: -1 },
                active: { line: 3, character: -1 },
                isReversed: false,
            },
        ]
        editorService.setSelections(editor, SELECTIONS)
        expect(
            await from(editorService.editors)
                .pipe(
                    first(),
                    map(editors => editors.map(e => e.selections))
                )
                .toPromise()
        ).toEqual([SELECTIONS])
    })

    test('removeEditor', async () => {
        const editorService = createEditorService({ models: of([{ uri: 'u', text: 't', languageId: 'l' }]) })
        const editor = editorService.addEditor({ type: 'CodeEditor', resource: 'u', selections: [], isActive: true })
        editorService.removeEditor(editor)
        expect(
            await from(editorService.editors)
                .pipe(first())
                .toPromise()
        ).toEqual([])
    })

    test('removeAllEditors', async () => {
        const editorService = createEditorService({ models: of([{ uri: 'u', text: 't', languageId: 'l' }]) })
        editorService.addEditor({ type: 'CodeEditor', resource: 'u', selections: [], isActive: true })
        editorService.removeAllEditors()
        expect(
            await from(editorService.editors)
                .pipe(first())
                .toPromise()
        ).toEqual([])
    })
})

describe('getActiveCodeEditorPosition', () => {
    test('null if code editor is empty', () => {
        expect(getActiveCodeEditorPosition([])).toBeNull()
    })

    test('null if no code editors are active', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'CodeEditor',
                    isActive: false,
                    selections: [],
                    resource: 'u',
                },
            ])
        ).toBeNull()
    })

    test('null if active code editor has no selection', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'CodeEditor',
                    isActive: true,
                    selections: [],
                    resource: 'u',
                },
            ])
        ).toBeNull()
    })

    test('null if active code editor has empty selection', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'CodeEditor',
                    isActive: true,
                    selections: [
                        {
                            start: { line: 3, character: -1 },
                            end: { line: 3, character: -1 },
                            anchor: { line: 3, character: -1 },
                            active: { line: 3, character: -1 },
                            isReversed: false,
                        },
                    ],
                    resource: 'u',
                },
            ])
        ).toBeNull()
    })

    test('equivalent params', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'CodeEditor',
                    isActive: true,
                    selections: [
                        {
                            start: { line: 3, character: 2 },
                            end: { line: 3, character: 5 },
                            anchor: { line: 3, character: 2 },
                            active: { line: 3, character: 5 },
                            isReversed: false,
                        },
                    ],
                    resource: 'u',
                },
            ])
        ).toEqual({ textDocument: { uri: 'u' }, position: { line: 3, character: 2 } })
    })
})
