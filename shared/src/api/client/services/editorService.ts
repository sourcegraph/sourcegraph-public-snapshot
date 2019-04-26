import { Selection } from '@sourcegraph/extension-api-types'
import { BehaviorSubject, combineLatest, Subscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import { TextDocumentPositionParams } from '../../protocol'
import { ModelService, TextModel } from './modelService'
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
export interface CodeEditor extends EditorId, CodeEditorData {
    /**
     * The model that represents the editor's document (and includes its contents).
     */
    readonly model: TextModel
}

/**
 * The editor service manages editors and documents.
 */
export interface EditorService {
    /** All code editors, with each editor's model. */
    readonly editors: Subscribable<readonly CodeEditor[]>

    /**
     * Add an editor.
     *
     * @param editor The description of the editor to add.
     * @returns The added code editor (which must be passed as the first argument to other
     * {@link EditorService} methods to operate on this editor).
     */
    addEditor(editor: CodeEditorData): EditorId

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
export function createEditorService(modelService: Pick<ModelService, 'models' | 'removeModel'>): EditorService {
    // Don't use lodash.uniqueId because that makes it harder to hard-code expected ID values in
    // test code (because the IDs change depending on test execution order).
    let id = 0
    const nextId = () => `editor#${id++}`

    const findModelForEditor = (models: readonly TextModel[], { resource }: Pick<CodeEditorData, 'resource'>) => {
        const model = models.find(m => m.uri === resource)
        if (!model) {
            throw new Error(`editor model not found: ${resource}`)
        }
        return model
    }

    type AddedCodeEditor = Pick<CodeEditor, Exclude<keyof CodeEditor, 'model'>>
    const editors = new BehaviorSubject<readonly AddedCodeEditor[]>([])
    const getEditor = (editorId: EditorId['editorId']) => editors.value.find(e => e.editorId === editorId)
    const exists = (editorId: EditorId['editorId']) => !!getEditor(editorId)
    return {
        editors: combineLatest(editors, modelService.models).pipe(
            map(([editors, models]) =>
                editors.map(editor => ({ ...editor, model: findModelForEditor(models, editor) }))
            )
        ),
        addEditor: data => {
            const editor: AddedCodeEditor = { ...data, editorId: nextId() }
            editors.next([...editors.value, editor])
            return editor
        },
        hasEditor: ({ editorId }) => exists(editorId),
        setSelections({ editorId }: EditorId, selections: Selection[]): void {
            if (!exists(editorId)) {
                throw new Error(`editor not found: ${editorId}`)
            }
            editors.next([
                ...editors.value.filter(e => e.editorId !== editorId),
                { ...editors.value.find(e => e.editorId === editorId)!, selections },
            ])
        },
        removeEditor({ editorId }: EditorId): void {
            const editor = getEditor(editorId)
            if (!editor) {
                throw new Error(`editor not found: ${editorId}`)
            }
            const nextEditors = editors.value.filter(e => e.editorId !== editorId)
            editors.next(editors.value.filter(e => e.editorId !== editorId))
            // If no other editor points to the same resouce,
            // remove the resource.
            if (!nextEditors.some(e => e.resource === editor.resource)) {
                modelService.removeModel(editor.resource)
            }
        },
        removeAllEditors(): void {
            editors.next([])
        },
    }
}

/**
 * Helper function to get the active editor's {@link TextDocumentPositionParams} from
 * {@link EditorService#editors}. If there is no active editor or it has no position, it returns
 * null.
 */
export function getActiveCodeEditorPosition(editors: readonly CodeEditorData[]): TextDocumentPositionParams | null {
    const activeEditor = editors.find(({ isActive }) => isActive)
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
