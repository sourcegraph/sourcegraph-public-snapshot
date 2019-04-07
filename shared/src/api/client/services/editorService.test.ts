import { from, of, Subscribable } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import {
    CodeEditorDataWithModel,
    createEditorService,
    EditorService,
    getActiveCodeEditorPosition,
} from './editorService'
import { TextModel } from './modelService'

export function createTestEditorService(
    editorsWithModel: Subscribable<readonly CodeEditorDataWithModel[]> = of([])
): Pick<EditorService, 'editors' | 'editorsWithModel'> {
    return { editors: editorsWithModel, editorsWithModel }
}

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('EditorService', () => {
    describe('editorsWithModel', () => {
        test('merges in model data', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const editorService = createEditorService({
                    models: cold<TextModel[]>('a', { a: [{ uri: 'u', text: 't', languageId: 'l' }] }),
                })
                editorService.nextEditors([{ type: 'CodeEditor', resource: 'u', selections: [], isActive: true }])
                expectObservable(from(editorService.editorsWithModel)).toBe('a', {
                    a: [
                        {
                            type: 'CodeEditor',
                            resource: 'u',
                            model: { uri: 'u', text: 't', languageId: 'l' },
                            selections: [],
                            isActive: true,
                        },
                    ],
                })
            })
        })
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
