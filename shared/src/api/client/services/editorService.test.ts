import { Selection } from '@sourcegraph/extension-api-types'
import { from, Observable } from 'rxjs'
import { first, tap } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import {
    createEditorService,
    EditorService,
    getActiveCodeEditorPosition,
    CodeEditorData,
    EditorUpdate,
} from './editorService'
import { createTestModelService } from './modelService.test'
import { ModelService } from './modelService'

export function createTestEditorService({
    modelService,
    editors,
    updates,
}: {
    modelService?: ModelService
    editors?: CodeEditorData[]
    updates?: Observable<EditorUpdate[]>
}): EditorService {
    const editorService = createEditorService(modelService || createTestModelService({}))
    if (editors) {
        for (const e of editors) {
            editorService.addEditor(e)
        }
    }
    const editorUpdates = updates
        ? from(updates).pipe(
              tap(updates => {
                  for (const update of updates) {
                      switch (update.type) {
                          case 'added':
                              editorService.addEditor(update.data)
                              break
                          case 'updated':
                              editorService.setSelections(update, update.data.selections)
                              break
                          case 'deleted':
                              editorService.removeEditor(update)
                              break
                      }
                  }
              })
          )
        : editorService.editorUpdates
    return { ...editorService, editorUpdates }
}

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('EditorService', () => {
    test('editors', () => {
        scheduler().run(({ expectObservable }) => {
            const editorService = createEditorService(
                createTestModelService({
                    models: [{ uri: 'u', text: 't', languageId: 'l' }],
                })
            )
            const editor: CodeEditorData = {
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            }
            const { editorId } = editorService.addEditor(editor)
            expectObservable(from(editorService.editorUpdates)).toBe('a', {
                a: [{ type: 'added', editorId, data: editor }],
            })
        })
    })

    // describe('editorsAndModels', () => {
    //     test('emits on model language change but not text change', () => {
    //         scheduler().run(({ cold, expectObservable }) => {
    //             const editorService = createEditorService(
    //                 createTestModelService({
    //                     models: [{ uri: 'u', text: 't', languageId: 'l' }],
    //                     updates: cold<TextModelUpdate[]>('-bc', {
    //                         b: [{ type: 'updated', uri: 'u', text: 't2', languageId: 'l' }],
    //                         c: [{ type: 'updated', uri: 'u', text: 't2', languageId: 'l2' }],
    //                     }),
    //                 })
    //             )
    //             const { editorId } = editorService.addEditor({
    //                 type: 'CodeEditor',
    //                 resource: 'u',
    //                 selections: [],
    //                 isActive: true,
    //             })
    //             expectObservable(from(editorService.editorUpdates)).toBe('a-c', {
    //                 a: [
    //                     {
    //                         type: 'added',
    //                         editorId,
    //                         data: {
    //                             type: 'CodeEditor',
    //                             resource: 'u',
    //                             selections: [],
    //                             isActive: true,
    //                             model: { uri: 'u', text: 't', languageId: 'l' },
    //                         },
    //                     },
    //                 ],
    //                 c: [
    //                     {
    //                         type: 'added',
    //                         editorId,
    //                         data: {
    //                             type: 'CodeEditor',
    //                             resource: 'u',
    //                             selections: [],
    //                             isActive: true,
    //                             model: { uri: 'u', text: 't2', languageId: 'l2' },
    //                         },
    //                     },
    //                 ],
    //             })
    //         })
    //     })
    // })

    test('addEditor', async () => {
        const editorService = createEditorService(
            createTestModelService({
                models: [{ uri: 'u', text: 't', languageId: 'l' }],
            })
        )
        const editorData: CodeEditorData = {
            type: 'CodeEditor',
            resource: 'u',
            selections: [],
            isActive: true,
        }
        const { editorId } = editorService.addEditor(editorData)
        expect(editorId).toEqual('editor#0')
        expect(
            await from(editorService.editorUpdates)
                .pipe(first())
                .toPromise()
        ).toEqual([
            {
                type: 'added',
                editorId,
                data: editorData,
            },
        ])
    })

    describe('observeEditorAndModel', () => {
        test('merges in model', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const editorService = createEditorService(
                    createTestModelService({
                        models: [{ uri: 'u', text: 't', languageId: 'l' }],
                        updates: cold('-b', {
                            b: [{ type: 'updated', uri: 'u', text: 't2' }],
                        }),
                    })
                )
                const editor: CodeEditorData = {
                    type: 'CodeEditor',
                    resource: 'u',
                    selections: [],
                    isActive: true,
                }
                const { editorId } = editorService.addEditor(editor)
                expectObservable(from(editorService.observeEditorAndModel({ editorId }))).toBe('ab', {
                    a: {
                        editorId,
                        ...editor,
                        model: { uri: 'u', text: 't', languageId: 'l' },
                    },
                    b: {
                        editorId,
                        ...editor,
                        model: { uri: 'u', text: 't2', languageId: 'l' },
                    },
                })
            })
        })

        test('emits error if editor does not exist', () => {
            scheduler().run(({ expectObservable }) => {
                const editorService = createEditorService(createTestModelService({}))
                expectObservable(from(editorService.observeEditorAndModel({ editorId: 'x' }))).toBe(
                    '#',
                    {},
                    new Error('editor not found: x')
                )
            })
        })
    })

    describe('setSelections', () => {
        test('ok', () => {
            scheduler().run(({ expectObservable }) => {
                const editorService = createEditorService(
                    createTestModelService({
                        models: [{ uri: 'u', text: 't', languageId: 'l' }],
                    })
                )
                const editor: CodeEditorData = {
                    type: 'CodeEditor',
                    resource: 'u',
                    selections: [],
                    isActive: true,
                }
                const { editorId } = editorService.addEditor(editor)
                const SELECTIONS: Selection[] = [
                    {
                        start: { line: 3, character: -1 },
                        end: { line: 3, character: -1 },
                        anchor: { line: 3, character: -1 },
                        active: { line: 3, character: -1 },
                        isReversed: false,
                    },
                ]
                editorService.setSelections({ editorId }, SELECTIONS)
                expectObservable(from(editorService.editorUpdates)).toBe('a', {
                    a: [{ type: 'updated', editorId, data: { selections: SELECTIONS } }],
                })
            })
        })
        test('not found', () => {
            const editorService = createEditorService(createTestModelService({}))
            expect(() => editorService.setSelections({ editorId: 'x' }, [])).toThrowError('editor not found: x')
        })
    })

    describe('removeEditor', () => {
        test('ok', () => {
            const editorService = createEditorService(
                createTestModelService({
                    models: [{ uri: 'u', text: 't', languageId: 'l' }],
                })
            )
            const editor = editorService.addEditor({
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            })
            editorService.removeEditor(editor)
            expect(editorService.editors.size).toBe(0)
        })
        test('not found', () => {
            const editorService = createEditorService(createTestModelService({}))
            expect(() => editorService.removeEditor({ editorId: 'x' })).toThrowError('editor not found: x')
        })

        it('calls removeModel() when removing the last editor that references the model', () => {
            const modelService = createTestModelService({
                models: [{ uri: 'u', text: 't', languageId: 'l' }],
            })
            const editorService = createEditorService(modelService)
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
            expect(modelService.hasModel('u')).toBeTruthy()
            editorService.removeEditor(editor2)
            expect(modelService.hasModel('u')).toBeFalsy()
        })
    })

    test('removeAllEditors', () => {
        scheduler().run(({ expectObservable }) => {
            const editorService = createEditorService(
                createTestModelService({
                    models: [{ uri: 'u', text: 't', languageId: 'l' }],
                })
            )
            const editor: CodeEditorData = { type: 'CodeEditor', resource: 'u', selections: [], isActive: true }
            editorService.addEditor(editor)
            editorService.addEditor(editor)
            editorService.addEditor(editor)
            editorService.removeAllEditors()
            expectObservable(from(editorService.editorUpdates)).toBe('a', {
                a: [
                    { type: 'deleted', editorId: 'editor#0' },
                    { type: 'deleted', editorId: 'editor#1' },
                    { type: 'deleted', editorId: 'editor#2' },
                ],
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
