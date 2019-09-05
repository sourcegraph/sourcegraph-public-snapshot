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
                              editorService.addEditor(update.editorData)
                              break
                          case 'updated':
                              editorService.setSelections(update, update.editorData.selections)
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
        const editorAdded = from(editorService.editorUpdates)
            .pipe(first())
            .toPromise()
        const { editorId } = editorService.addEditor(editorData)
        expect(editorId).toEqual('editor#0')
        expect(await editorAdded).toEqual([
            {
                type: 'added',
                editorId,
                editorData,
            },
        ] as EditorUpdate[])
    })

    describe('observeEditorAndModel', () => {
        test('merges in model', async () => {
            const modelService = createTestModelService({
                models: [{ uri: 'u', text: 't', languageId: 'l' }],
            })
            const editorService = createEditorService(modelService)
            const editor: CodeEditorData = {
                type: 'CodeEditor',
                resource: 'u',
                selections: [],
                isActive: true,
            }
            const { editorId } = editorService.addEditor(editor)
            expect(
                await editorService
                    .observeEditorAndModel({ editorId })
                    .pipe(first())
                    .toPromise()
            ).toMatchObject({
                editorId,
                ...editor,
                model: { uri: 'u', text: 't', languageId: 'l' },
            })
            modelService.updateModel('u', 't2')
            expect(
                await editorService
                    .observeEditorAndModel({ editorId })
                    .pipe(first())
                    .toPromise()
            ).toMatchObject({
                editorId,
                ...editor,
                model: { uri: 'u', text: 't2', languageId: 'l' },
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
        test('ok', async () => {
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
            const selectionsSet = from(editorService.editorUpdates)
                .pipe(first())
                .toPromise()
            editorService.setSelections({ editorId }, SELECTIONS)
            expect(await selectionsSet).toMatchObject([
                { type: 'updated', editorId, editorData: { selections: SELECTIONS } },
            ] as EditorUpdate[])
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

        it('calls removeModelRef() when removing an editor', () => {
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

    test('removeAllEditors', async () => {
        const editorService = createEditorService(
            createTestModelService({
                models: [{ uri: 'u', text: 't', languageId: 'l' }],
            })
        )
        const editor: CodeEditorData = { type: 'CodeEditor', resource: 'u', selections: [], isActive: true }
        editorService.addEditor(editor)
        editorService.addEditor(editor)
        editorService.addEditor(editor)
        const editorsRemoved = from(editorService.editorUpdates)
            .pipe(first())
            .toPromise()
        editorService.removeAllEditors()
        expect(await editorsRemoved).toMatchObject([
            { type: 'deleted', editorId: 'editor#0' },
            { type: 'deleted', editorId: 'editor#1' },
            { type: 'deleted', editorId: 'editor#2' },
        ])
    })
})

describe('getActiveCodeEditorPosition', () => {
    test('null if code editor is empty', () => {
        expect(getActiveCodeEditorPosition(undefined)).toBeNull()
    })

    test('null if active code editor has no selection', () => {
        expect(
            getActiveCodeEditorPosition({
                editorId: '1',
                type: 'CodeEditor',
                isActive: true,
                selections: [],
                resource: 'u',
            })
        ).toBeNull()
    })

    test('null if active code editor has empty selection', () => {
        expect(
            getActiveCodeEditorPosition({
                editorId: '1',
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
            })
        ).toBeNull()
    })

    test('equivalent params', () => {
        expect(
            getActiveCodeEditorPosition({
                editorId: '1',
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
            })
        ).toEqual({ textDocument: { uri: 'u' }, position: { line: 3, character: 2 } })
    })
})
