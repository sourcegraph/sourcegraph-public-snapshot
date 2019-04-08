import { Selection } from '@sourcegraph/extension-api-types'
import { noop } from 'lodash'
import { from, of, Subscribable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import * as sinon from 'sinon'
import {
    CodeEditor,
    createEditorService,
    EditorService,
    getActiveCodeEditorPosition,
    getEditorModels,
} from './editorService'
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
                    models: cold<TextModel[]>('a', {
                        a: [{ uri: 'u', text: 't', languageId: 'l' }, { uri: 'u2', text: 't2', languageId: 'l2' }],
                    }),
                    removeModel: noop,
                })
                editorService.addEditor({ type: 'CodeEditor', resource: 'u', selections: [], isActive: true })
                editorService.addEditor({
                    type: 'DiffEditor',
                    originalResource: 'u',
                    modifiedResource: 'u2',
                    rawDiff: 'x',
                    isActive: true,
                })
                expectObservable(
                    from(editorService.editors).pipe(
                        map(editors =>
                            editors.map(e => {
                                switch (e.type) {
                                    case 'CodeEditor':
                                        return { resource: e.resource, model: e.model }
                                    case 'DiffEditor':
                                        return {
                                            originalResource: e.originalResource,
                                            originalModel: e.originalModel,
                                            modifiedResource: e.modifiedResource,
                                            modifiedModel: e.modifiedModel,
                                        }
                                }
                            })
                        )
                    )
                ).toBe('a', {
                    a: [
                        {
                            resource: 'u',
                            model: { uri: 'u', text: 't', languageId: 'l' },
                        },
                        {
                            originalResource: 'u',
                            originalModel: { uri: 'u', text: 't', languageId: 'l' },
                            modifiedResource: 'u2',
                            modifiedModel: { uri: 'u2', text: 't2', languageId: 'l2' },
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
            const editors = await from(editorService.editors)
                .pipe(first())
                .toPromise()
            const selections = editors.filter((e): e is CodeEditor => e.type === 'CodeEditor').map(e => e.selections)
            expect(selections).toEqual([SELECTIONS])
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

    test('null for diff editor', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'DiffEditor',
                    isActive: true,
                    originalResource: 'u',
                    modifiedResource: 'u',
                    rawDiff: 'x',
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

describe('getEditorModels', () => {
    test('CodeEditor', () =>
        expect(
            getEditorModels({
                type: 'CodeEditor',
                editorId: 'editor#0',
                isActive: true,
                selections: [],
                resource: 'u',
                model: { uri: 'u', text: 't', languageId: 'l' },
            })
        ).toEqual([{ uri: 'u', text: 't', languageId: 'l' }]))

    test('DiffEditor', () =>
        expect(
            getEditorModels({
                type: 'DiffEditor',
                editorId: 'editor#0',
                isActive: true,
                rawDiff: 'x',
                originalResource: 'u',
                originalModel: { uri: 'u', text: 't', languageId: 'l' },
                modifiedResource: 'u2',
                modifiedModel: { uri: 'u2', text: 't2', languageId: 'l2' },
            })
        ).toEqual([{ uri: 'u', text: 't', languageId: 'l' }, { uri: 'u2', text: 't2', languageId: 'l2' }]))
})
