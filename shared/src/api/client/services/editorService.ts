import { Selection } from '@sourcegraph/extension-api-types'
import { BehaviorSubject, Subscribable } from 'rxjs'
import { map, publishReplay, refCount, withLatestFrom } from 'rxjs/operators'
import { Omit } from 'utility-types'
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
 * EditorId exposes the unique ID of an editor.
 */
export interface EditorId {
    /** The unique ID of the editor. */
    readonly editorId: string
}

export interface EditorDataCommon {
    readonly collapsed?: boolean
}

/**
 * Describes a code editor to be created.
 */
export interface CodeEditorData extends EditorDataCommon {
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
 * Describes a diff editor to be created.
 */
export interface DiffEditorData extends EditorDataCommon {
    readonly type: 'DiffEditor'

    /** The URI of the left-hand side resource. */
    readonly originalResource: string

    /** The URI of the right-hand side resource. */
    readonly modifiedResource: string

    /** The raw unified diff. */
    readonly rawDiff: string | undefined

    readonly isActive: boolean
}

/**
 * Describes a code editor that has been added to the {@link EditorService}.
 */
export interface DiffEditor extends EditorId, DiffEditorData {
    /**
     * The model that represents the original document (on the left-hand side).
     */
    readonly originalModel: TextModel

    /**
     * The model that represents the modified document (on the right-hand side).
     */
    readonly modifiedModel: TextModel
}

type EditorData = CodeEditorData | DiffEditorData

/** All editor types. */
export type Editor = CodeEditor | DiffEditor

/**
 * The editor service manages editors and documents.
 */
export interface EditorService {
    /**
     * All code editors, with each editor's model.
     * Emits the current value upon subscription.
     */
    readonly editors: Subscribable<readonly (CodeEditor | DiffEditor)[]>

    /**
     * Add an editor.
     *
     * @param editor The description of the editor to add.
     * @returns The added editor ID (which must be passed as the first argument to other
     * {@link EditorService} methods to operate on this editor).
     */
    addEditor(editor: EditorData): EditorId

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

    setCollapsed({ editorId }: EditorId, collapsed: boolean): void
}

type EditorWithoutModels = Omit<CodeEditor, 'model'> | Omit<DiffEditor, 'originalModel' | 'modifiedModel'>

/**
 * Creates a {@link EditorService} instance.
 */
export function createEditorService(modelService: Pick<ModelService, 'models' | 'removeModel'>): EditorService {
    // Don't use lodash.uniqueId because that makes it harder to hard-code expected ID values in
    // test code (because the IDs change depending on test execution order).
    let id = 0
    const nextId = () => `editor#${id++}`

    const findModelForEditor = (models: readonly TextModel[], uri: string): TextModel => {
        const model = models.find(m => m.uri === uri)
        if (!model) {
            throw new Error(`editor model not found: ${uri}`)
        }
        return model
    }

    const editors = new BehaviorSubject<readonly EditorWithoutModels[]>([])
    const getEditor = (editorId: EditorId['editorId']) => editors.value.find(e => e.editorId === editorId)
    const exists = (editorId: EditorId['editorId']) => !!getEditor(editorId)
    return {
        editors: editors.pipe(
            withLatestFrom(modelService.models), // Do not emit on model changes
            map(([editors, models]) =>
                editors.map(
                    (editor): CodeEditor | DiffEditor => {
                        switch (editor.type) {
                            case 'CodeEditor':
                                return { ...editor, model: findModelForEditor(models, editor.resource) }
                            case 'DiffEditor':
                                return {
                                    ...editor,
                                    originalModel: findModelForEditor(models, editor.originalResource),
                                    modifiedModel: findModelForEditor(models, editor.modifiedResource),
                                }
                        }
                    }
                )
            ),
            // Perf optimization: avoid running findModelForEditor() for every subscriber
            // This does not change the behaviour of the Observable.
            publishReplay(1),
            refCount()
        ),
        addEditor: data => {
            const editorId = nextId()
            const editor: EditorWithoutModels = { ...data, editorId }
            editors.next([...editors.value, editor])
            return { editorId }
        },
        hasEditor: ({ editorId }) => exists(editorId),
        setSelections({ editorId }: EditorId, selections: Selection[]): void {
            const editor = getEditor(editorId)
            if (!editor) {
                throw new Error(`Editor not found: ${editorId}`)
            }
            if (editor.type !== 'CodeEditor') {
                throw new Error(`${editor.type}s do not have selections`)
            }
            editors.next([...editors.value.filter(e => e.editorId !== editorId), { ...editor, selections }])
        },
        removeEditor({ editorId }: EditorId): void {
            const editor = getEditor(editorId)
            if (!editor) {
                throw new Error(`Editor not found: ${editorId}`)
            }
            const nextEditors = editors.value.filter(e => e.editorId !== editorId)
            editors.next(nextEditors)
            // If no other editor points to a resource held by this editor, remove it
            const stillReferencedResources = new Set(nextEditors.flatMap(getEditorResources))
            for (const editorResource of getEditorResources(editor)) {
                if (!stillReferencedResources.has(editorResource)) {
                    modelService.removeModel(editorResource)
                }
            }
        },
        removeAllEditors(): void {
            editors.next([])
        },
        setCollapsed({ editorId }: EditorId, collapsed: boolean): void {
            const editor = getEditor(editorId)
            if (!editor) {
                throw new Error(`Editor not found: ${editorId}`)
            }
            editors.next([...editors.value.filter(e => e.editorId !== editorId), { ...editor, collapsed }])
        },
    }
}

/**
 * Helper function to get the active editor's {@link TextDocumentPositionParams} from
 * {@link EditorService#editors}. If there is no active editor or it has no position, it returns
 * null.
 */
export function getActiveCodeEditorPosition(editors: readonly EditorData[]): TextDocumentPositionParams | null {
    const activeEditor = editors.find(({ isActive }) => isActive)
    if (!activeEditor) {
        return null
    }
    if (activeEditor.type !== 'CodeEditor') {
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
 * Returns an array of all models referenced by the editor. A code editor has 1 model; a diff editor
 * has 2 models (original/modified, also known as the left-hand and right-hand sides).
 */
export function getEditorModels(editor: Editor): TextModel[] {
    switch (editor.type) {
        case 'CodeEditor':
            return [editor.model]
        case 'DiffEditor':
            return [editor.originalModel, editor.modifiedModel]
    }
}

/**
 * Returns an array of all resources (URIs) referenced by the editor. A code editor has 1 model; a diff editor
 * has 2 models (original/modified, also known as the left-hand and right-hand sides).
 */
function getEditorResources(editor: EditorWithoutModels): string[] {
    switch (editor.type) {
        case 'CodeEditor':
            return [editor.resource]
        case 'DiffEditor':
            return [editor.originalResource, editor.modifiedResource]
    }
}
