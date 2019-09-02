import { Selection } from '@sourcegraph/extension-api-types'
import { isEqual } from 'lodash'
import { BehaviorSubject, combineLatest, from, Subscribable, throwError, zip, Subject, NEVER } from 'rxjs'
import { distinctUntilChanged, map, switchMap, filter, takeUntil, takeWhile, startWith } from 'rxjs/operators'
import { TextDocumentPositionParams } from '../../protocol'
import { ModelService, TextModel, PartialModel } from './modelService'
/**
 * EditorId exposes the unique ID of an editor.
 */
export interface EditorId {
    /** The unique ID of the editor. */
    readonly editorId: string
}

/**
 * Describes a code editor to be created.
 */
export interface CodeEditorData {
    readonly type: 'CodeEditor'

    /** The URI of the model that this editor is displaying. */
    readonly resource: string

    readonly selections: Selection[]
    readonly isActive: boolean
}

/**
 * Describes a code editor that has been added to the {@link EditorService}. To get the editor's
 * model, use {@link EditorService#observeEditorAndModel} or look up the model in the
 * {@link ModelService}.
 */
export interface CodeEditor extends EditorId, CodeEditorData {}

/**
 * A code editor with some fields from its model.
 */
export interface CodeEditorWithPartialModel extends CodeEditor {
    /**
     * The code editor's immutable model data.
     */
    model: PartialModel
}

/**
 * A code editor with its full model, including the model text.
 */
export interface CodeEditorWithModel extends CodeEditor {
    /** The code editor's model. */
    model: TextModel
}

export type EditorUpdate =
    | { type: 'added'; data: CodeEditorData } & EditorId
    | { type: 'updated'; data: Pick<CodeEditor, 'selections'> } & EditorId
    | { type: 'deleted' } & EditorId

/**
 * The editor service manages editors and documents.
 */
export interface EditorService {
    /**
     * All code editors.
     */
    readonly editors: ReadonlyMap<string, CodeEditorWithPartialModel>

    readonly editorUpdates: Subscribable<EditorUpdate[]>

    readonly activeEditorUpdates: Subscribable<CodeEditorWithPartialModel | undefined>

    readonly activeLanguages: Subscribable<string[]>

    /**
     * Add an editor.
     *
     * @param editor The description of the editor to add.
     * @returns The added code editor (which must be passed as the first argument to other
     * {@link EditorService} methods to operate on this editor).
     */
    addEditor(editor: CodeEditorData): EditorId

    /**
     * Observe an editor and its model for changes, including changes to the model's content.
     *
     * @param editor The editor to observe.
     * @returns An observable that emits when the editor or its model changes. If no such editor
     * exists, or if the editor is removed, it emits an error.
     */
    observeEditorAndModel(editor: EditorId): Subscribable<CodeEditorWithModel>

    /**
     * Reports whether an editor with the given URI has been added.
     *
     * @param editor the {@link EditorId} to check
     */
    hasEditor(editor: EditorId): boolean

    /**
     * Sets the selections for an editor.
     *
     * @param editor The editor for which to set the selections.
     * @param selections The new selections to apply.
     * @throws if no editor exists with the given editor ID.
     */
    setSelections(editor: EditorId, selections: Selection[]): void

    /**
     * Removes an editor.
     * Also removes the corresponding model if no other editor is referencing it.
     *
     * @param editor The editor to remove.
     */
    removeEditor(editor: EditorId): void

    /**
     * Remove all editors.
     */
    removeAllEditors(): void
}

/**
 * Creates a {@link EditorService} instance.
 */
export function createEditorService(
    modelService: Pick<ModelService, 'models' | 'modelUpdates' | 'removeModel'>
): EditorService {
    // Don't use lodash.uniqueId because that makes it harder to hard-code expected ID values in
    // test code (because the IDs change depending on test execution order).
    let id = 0
    const nextId = (): string => `editor#${id++}`

    /** A map of editor ids to code editors. */
    const editors = new Map<string, CodeEditorWithPartialModel>()
    const editorUpdates = new BehaviorSubject<EditorUpdate[]>([])
    const activeEditorUpdates = new BehaviorSubject<CodeEditorWithPartialModel | undefined>(undefined)
    const activeLanguagesSet = new Set<string>()
    const activeLanguages = new BehaviorSubject<string[]>([])
    /**
     * Returns the CodeEditor with the given editorId.
     * Throws if no editor exists with the given editorId.
     */
    const getEditor = (editorId: EditorId['editorId']): CodeEditorWithPartialModel => {
        const editor = editors.get(editorId)
        if (!editor) {
            throw new Error(`editor not found: ${editorId}`)
        }
        return editor
    }
    return {
        editors,
        editorUpdates,
        activeEditorUpdates,
        activeLanguages,
        addEditor: data => {
            const editorId = nextId()
            const editor: CodeEditorWithPartialModel = {
                ...data,
                editorId,
                model: {
                    languageId: modelService.models.get(data.resource)!.languageId,
                },
            }
            editors.set(editorId, editor)
            editorUpdates.next([{ type: 'added', editorId, data }])
            if (data.isActive) {
                activeEditorUpdates.next(editor)
            }
            if (!activeLanguagesSet.has(editor.model.languageId)) {
                activeLanguagesSet.add(editor.model.languageId)
                activeLanguages.next([...activeLanguagesSet])
            }
            return { editorId }
        },
        observeEditorAndModel: ({ editorId }) => {
            let resource: string
            try {
                resource = getEditor(editorId).resource
            } catch (err) {
                return throwError(err)
            }
            return combineLatest(
                editorUpdates.pipe(
                    filter(updates => updates.some(u => u.editorId === editorId)),
                    takeWhile(updates => updates.every(u => u.editorId !== editorId || u.type !== 'deleted')),
                    map(() => getEditor(editorId)),
                    startWith(getEditor(editorId))
                ),
                from(modelService.modelUpdates).pipe(
                    filter(updates => updates.some(u => u.uri === resource)),
                    takeWhile(updates => updates.every(u => u.uri !== resource || u.type !== 'deleted')),
                    map(() => modelService.models.get(resource)!),
                    startWith(modelService.models.get(resource)!)
                )
            ).pipe(
                map(([editor, model]) => ({
                    ...editor,
                    model,
                }))
            )
        },
        hasEditor: ({ editorId }) => editors.has(editorId),
        setSelections({ editorId }: EditorId, selections: Selection[]): void {
            const editor = getEditor(editorId)
            editors.set(editorId, {
                ...editor,
                selections,
            })
            editorUpdates.next([{ type: 'updated', editorId, data: { selections } }])
        },
        removeEditor({ editorId }: EditorId): void {
            const editor = getEditor(editorId)
            editors.delete(editorId)
            editorUpdates.next([{ type: 'deleted', editorId }])
            // Check if this was the active editor
            if (activeEditorUpdates.value && activeEditorUpdates.value.editorId === editorId) {
                activeEditorUpdates.next(undefined)
            }
            // Check if any of the remaining editors points to the same resource
            // or has the same language
            let resourceMatch = false
            let languageMatch = false
            for (const e of editors.values()) {
                if (e.resource === editor.resource) {
                    resourceMatch = true
                }
                if (e.model.languageId === editor.model.languageId) {
                    languageMatch = true
                }
                if (resourceMatch && languageMatch) {
                    return
                }
            }
            // If no other editor points to the same resource, remove the resource.
            if (!resourceMatch) {
                modelService.removeModel(editor.resource)
            }
            // If no other editor has the same language, update activeLanguages
            if (!languageMatch) {
                activeLanguagesSet.delete(editor.model.languageId)
                activeLanguages.next([...activeLanguagesSet])
            }
        },
        removeAllEditors(): void {
            const updates: EditorUpdate[] = [...editors.keys()].map(editorId => ({ type: 'deleted', editorId }))
            editors.clear()
            editorUpdates.next(updates)
        },
    }
}

/**
 * Helper function to get the active editor's {@link TextDocumentPositionParams} from
 * {@link EditorService#editors}. If there is no active editor or it has no position, it returns
 * null.
 */
export function getActiveCodeEditorPosition(
    activeEditor: CodeEditorWithPartialModel | undefined
): TextDocumentPositionParams | null {
    if (!activeEditor) {
        return null
    }
    const sel = activeEditor.selections[0]
    if (!sel) {
        return null
    }
    // TODO(sqs): Return null for empty selections (but currently all selected tokens are treated as an empty
    // selection at the beginning of the token, so this would break a lot of things, so we only do this for empty
    // selections when the start character is -1). HACK(sqs): Character === -1 means that the whole line is
    // selected (this is a bug in the caller, but it is useful here).
    const isEmpty =
        sel.start.line === sel.end.line && sel.start.character === sel.end.character && sel.start.character === -1
    if (isEmpty) {
        return null
    }
    return {
        textDocument: { uri: activeEditor.resource },
        position: sel.start,
    }
}
