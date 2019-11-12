import { Selection } from '@sourcegraph/extension-api-types'
import { BehaviorSubject, Subscribable, throwError, Observable, Subject } from 'rxjs'
import { map, filter, takeWhile, startWith, switchMap } from 'rxjs/operators'
import { TextDocumentPositionParams } from '../../protocol'
import { ModelService, TextModel, PartialModel } from './modelService'
import { RefCount } from '../../../util/RefCount'

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
 * Describes a code editor that has been added to the {@link EditorService}.
 */
export interface CodeEditor extends EditorId, CodeEditorData {}

/**
 * A code editor with a partial model.
 *
 * To get the editor's full model, look up the model in the {@link ModelService}.
 */
export interface CodeEditorWithPartialModel extends CodeEditor {
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
    | ({ type: 'added'; editorData: CodeEditorData } & EditorId)
    | ({ type: 'updated'; editorData: Pick<CodeEditorData, 'selections'> } & EditorId)
    | ({ type: 'deleted' } & EditorId)

/**
 * The editor service manages editors and documents.
 */
export interface EditorService {
    /**
     * A map of all known editors, indexed by editorId.
     *
     * This is mostly used for testing, most consumers should use
     * {@link EditorService#editorUpdates} or {@link EditorService#activeEditorUpdates}
     */
    readonly editors: ReadonlyMap<string, CodeEditor>

    /**
     * An observable of all editor updates.
     *
     * Emits when an editor is added, updated or removed.
     */
    readonly editorUpdates: Subscribable<EditorUpdate[]>

    /**
     * An observable of updates to the active editor.
     *
     * Emits the active editor if there is one, or `undefined` otherwise.
     */
    readonly activeEditorUpdates: Subscribable<CodeEditor | undefined>

    /**
     * Add an editor.
     *
     * @param editor The description of the editor to add.
     * @returns The added code editor (which must be passed as the first argument to other
     * {@link EditorService} methods to operate on this editor).
     */
    addEditor(editor: CodeEditorData): EditorId

    /**
     * Observe an editor for changes.
     *
     * @param editor The editor to observe.
     * @returns An observable that emits when the editor changes,
     * and completes when the editor is removed.
     * If no such editor exists, it emits an error.
     */
    observeEditor(editor: EditorId): Observable<CodeEditorData>

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

const EEDITORNOTFOUND = 'EditorNotFoundError'
class EditorNotFoundError extends Error {
    public readonly name = EEDITORNOTFOUND
    public readonly code = EEDITORNOTFOUND
    constructor(editorId: string) {
        super(`editor not found: ${editorId}`)
    }
}

/**
 * Creates a {@link EditorService} instance.
 */
export function createEditorService(modelService: Pick<ModelService, 'removeModel'>): EditorService {
    // Don't use lodash.uniqueId because that makes it harder to hard-code expected ID values in
    // test code (because the IDs change depending on test execution order).
    let id = 0
    const nextId = (): string => `editor#${id++}`

    /** A map of editor ids to code editors. */
    const editors = new Map<string, CodeEditor>()
    const editorUpdates = new Subject<EditorUpdate[]>()
    const activeEditorUpdates = new BehaviorSubject<CodeEditor | undefined>(undefined)
    /**
     * Returns the CodeEditor with the given editorId.
     * Throws if no editor exists with the given editorId.
     */
    const getEditor = (editorId: EditorId['editorId']): CodeEditor => {
        const editor = editors.get(editorId)
        if (!editor) {
            throw new EditorNotFoundError(editorId)
        }
        return editor
    }

    const modelRefs = new RefCount()
    return {
        editors,
        editorUpdates,
        activeEditorUpdates,
        addEditor: editorData => {
            const editorId = nextId()
            modelRefs.increment(editorData.resource)
            const editor: CodeEditor = {
                ...editorData,
                editorId,
            }
            editors.set(editorId, editor)
            editorUpdates.next([{ type: 'added', editorId, editorData }])
            if (editorData.isActive) {
                activeEditorUpdates.next(editor)
            }
            return { editorId }
        },
        observeEditor: ({ editorId }) => {
            try {
                const editor = getEditor(editorId)
                return editorUpdates.pipe(
                    filter(updates => updates.some(u => u.editorId === editorId)),
                    takeWhile(updates => updates.every(u => u.editorId !== editorId || u.type !== 'deleted')),
                    map(() => getEditor(editorId)),
                    startWith(editor)
                )
            } catch (err) {
                return throwError(err)
            }
        },
        setSelections({ editorId }: EditorId, selections: Selection[]): void {
            const editor = getEditor(editorId)
            editors.set(editorId, {
                ...editor,
                selections,
            })
            editorUpdates.next([{ type: 'updated', editorId, editorData: { selections } }])
        },
        removeEditor({ editorId }: EditorId): void {
            const editor = getEditor(editorId)
            editors.delete(editorId)
            editorUpdates.next([{ type: 'deleted', editorId }])
            // Check if this was the active editor
            if (activeEditorUpdates.value && activeEditorUpdates.value.editorId === editorId) {
                activeEditorUpdates.next(undefined)
            }
            if (modelRefs.decrement(editor.resource)) {
                modelService.removeModel(editor.resource)
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
export function getActiveCodeEditorPosition(activeEditor: CodeEditor | undefined): TextDocumentPositionParams | null {
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

/**
 * Observe an editor and its model for changes.
 */
export function observeEditorAndModel(
    { editorId }: EditorId,
    { observeEditor }: Pick<EditorService, 'observeEditor'>,
    { observeModel }: Pick<ModelService, 'observeModel'>
): Observable<CodeEditorWithModel> {
    return observeEditor({ editorId }).pipe(
        switchMap(editor => observeModel(editor.resource).pipe(map(model => ({ editorId, ...editor, model }))))
    )
}
