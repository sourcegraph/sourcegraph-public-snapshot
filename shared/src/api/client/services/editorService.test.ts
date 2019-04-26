import { Selection } from '@sourcegraph/extension-api-types'
import { noop } from 'lodash'
import { from, of, Subscribable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import * as sinon from 'sinon'
import { CodeEditor, createEditorService, EditorService, getActiveCodeEditorPosition } from './editorService'
import { ModelService, TextModel } from './modelService'

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
                    removeModel: jest.fn(),
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
        const editorService = createEditorService({
            models: of([{ uri: 'u', text: 't', languageId: 'l' }]),
            removeModel: jest.fn(),
        })
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

    describe('setSelections', () => {
        test('ok', async () => {
            const editorService = createEditorService({
                models: of([{ uri: 'u', text: 't', languageId: 'l' }]),
                removeModel: jest.fn(),
            })
            const editor = editorService.addEditor({
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            })
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
        test('not found', () => {
            const editorService = createEditorService({ models: of([]), removeModel: jest.fn() })
            expect(() => editorService.setSelections({ editorId: 'x' }, [])).toThrowError('editor not found: x')
        })
    })

    describe('removeEditor', () => {
        test('ok', async () => {
            const editorService = createEditorService({
                models: of([{ uri: 'u', text: 't', languageId: 'l' }]),
                removeModel: noop,
            })
            const editor = editorService.addEditor({
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            })
            editorService.removeEditor(editor)
            expect(
                await from(editorService.editors)
                    .pipe(first())
                    .toPromise()
            ).toEqual([])
        })
        test('not found', () => {
            const editorService = createEditorService({ models: of([]), removeModel: noop })
            expect(() => editorService.removeEditor({ editorId: 'x' })).toThrowError('editor not found: x')
        })

        it('calls removeModel() when removing the last editor that references the model', async () => {
            const removeModel = sinon.spy<ModelService['removeModel']>(noop)
            const editorService = createEditorService({
                models: of([{ uri: 'u', text: 't', languageId: 'l' }]),
                removeModel,
            })
            const editor1 = editorService.addEditor({
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            })
            const editor2 = editorService.addEditor({
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            })
            editorService.removeEditor(editor1)
            sinon.assert.notCalled(removeModel)
            editorService.removeEditor(editor2)
            expect(
                await from(editorService.editors)
                    .pipe(first())
                    .toPromise()
            ).toEqual([])
            sinon.assert.calledOnce(removeModel)
            sinon.assert.calledWith(removeModel, 'u')
        })
    })

    test('removeAllEditors', async () => {
        const editorService = createEditorService({
            models: of([{ uri: 'u', text: 't', languageId: 'l' }]),
            removeModel: noop,
        })
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
