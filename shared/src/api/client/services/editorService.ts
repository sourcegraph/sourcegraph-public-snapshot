import { Selection } from '@sourcegraph/extension-api-types'
import { isEqual } from 'lodash'
import { BehaviorSubject, combineLatest, from, Subscribable, throwError, of } from 'rxjs'
import { distinctUntilChanged, map, switchMap, first, filter, takeWhile } from 'rxjs/operators'
import { TextDocumentPositionParams } from '../../protocol'
import { ModelService, TextModel } from './modelService'
import { isDefined } from '../../../util/types'
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
     * The code editor's model. See {@link EditorService#editorsAndModels} for why the model text is
     * omitted. The model's URI is always equal to {@link CodeEditor#resource}, so it is also
     * omitted (to avoid confusion).
     */
    model: Pick<TextModel, 'languageId'>
}

/**
 * A code editor with its full model, including the model text.
 */
export interface CodeEditorWithModel extends CodeEditor {
    /** The code editor's model. */
    model: TextModel
}

/**
 * The editor service manages editors and documents.
 */
export interface EditorService {
    /**
     * All code editors. Emits the current value upon subscription and whenever any editor changes.
     */
    readonly editors: Subscribable<readonly CodeEditor[]>

    /**
     * All code editors, with each editor's model (minus its text). Emits the current value upon
     * subscription and whenever any editor or model changes.
     *
     * The model text is omitted for performance reasons and because there are not currently any
     * callers that need to observe all editors and their model text. Only the model's URI and
     * language are included.
     */
    readonly editorsAndModels: Subscribable<readonly CodeEditorWithPartialModel[]>

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
     * exists or if the editor is removed, the observable completes.
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
 * Picks from {@link CodeEditorWithModel} only the properties that should be considered to determine
 * if two editors are equal. It is useful because it removes the model's text from the object (which
 * is present on the object but not noted in the type).
 */
const pickCodeEditorWithModelEqualityData = ({
    editorId,
    type,
    isActive,
    resource,
    selections,
    model: { languageId },
}: CodeEditorWithPartialModel): CodeEditorWithPartialModel => ({
    editorId,
    type,
    resource,
    selections,
    isActive,
    model: { languageId },
})

/**
 * Creates a {@link EditorService} instance.
 */
export function createEditorService(modelService: Pick<ModelService, 'models' | 'removeModel'>): EditorService {
    // Don't use lodash.uniqueId because that makes it harder to hard-code expected ID values in
    // test code (because the IDs change depending on test execution order).
    let id = 0
    const nextId = (): string => `editor#${id++}`

    const findModelForEditor = (
        models: readonly TextModel[],
        { resource }: Pick<CodeEditorData, 'resource'>
    ): TextModel => {
        const model = models.find(m => m.uri === resource)
        if (!model) {
            // This error indicates the presence of a bug. The model should never be removed before
            // all editors using the model are removed.
            throw new Error(`editor model not found: ${resource}`)
        }
        return model
    }

    const editors = new BehaviorSubject<readonly CodeEditor[]>([])
    const getEditor = (editorId: EditorId['editorId']): CodeEditor | undefined =>
        editors.value.find(e => e.editorId === editorId)
    const exists = (editorId: EditorId['editorId']): boolean => !!getEditor(editorId)
    return {
        editors,
        editorsAndModels: combineLatest([editors, modelService.models]).pipe(
            map(([editors, models]) =>
                editors.map(editor => ({ ...editor, model: findModelForEditor(models, editor) }))
            ),
            distinctUntilChanged((a, b) =>
                isEqual(a.map(pickCodeEditorWithModelEqualityData), b.map(pickCodeEditorWithModelEqualityData))
            )
        ),
        addEditor: data => {
            const editorId = nextId()
            const editor: CodeEditor = { ...data, editorId }
            editors.next([...editors.value, editor])
            return { editorId }
        },
        observeEditorAndModel: ({ editorId }) =>
            combineLatest([
                editors.pipe(map(editors => editors.find(e => e.editorId === editorId))),
                from(modelService.models),
            ]).pipe(
                takeWhile((data): data is [CodeEditor, readonly TextModel[]] => isDefined(data[0])),
                map(([editor, models]) => ({ ...editor, model: findModelForEditor(models, editor) }))
            ),
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
            // If no other editor points to the same resource, remove the resource.
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
