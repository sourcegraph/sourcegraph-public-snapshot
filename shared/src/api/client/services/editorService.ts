import { Selection } from '@sourcegraph/extension-api-types'
import { BehaviorSubject, combineLatest, Subscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import { TextDocumentPositionParams } from '../../protocol'
import { ModelService, TextModel } from './modelService'

/**
 * Describes a code editor view component.
 */
export interface CodeEditorData {
    type: 'CodeEditor'

    /** The URI of the model that this editor is displaying. */
    resource: string

    selections: Selection[]
    isActive: boolean
}

/** Describes a code editor and includes its model content. */
export interface CodeEditorDataWithModel extends CodeEditorData {
    model: TextModel
}

/**
 * The editor service manages editors and documents.
 */
export interface EditorService {
    /** All code editors. */
    readonly editors: Subscribable<readonly CodeEditorData[]>

    /** All code editors, with each editor's model. */
    readonly editorsWithModel: Subscribable<readonly CodeEditorDataWithModel[]>

    /** Transitional API for synchronously getting the list of code editors. */
    readonly editorsValue: readonly CodeEditorData[]

    /** Transitional API for setting the list of code editors. */
    nextEditors(value: readonly CodeEditorData[]): void
}

/**
 * Creates a {@link EditorService} instance.
 */
export function createEditorService(modelService: Pick<ModelService, 'models'>): EditorService {
    const editors = new BehaviorSubject<readonly CodeEditorData[]>([])
    return {
        editors,
        editorsWithModel: combineLatest(editors, modelService.models).pipe(
            map(([editors, models]) =>
                editors.map(editor => {
                    const model = models.find(m => m.uri === editor.resource)
                    if (!model) {
                        throw new Error(`editor model not found: ${editor.resource}`)
                    }
                    return { ...editor, model }
                })
            )
        ),
        get editorsValue(): readonly CodeEditorData[] {
            return editors.value
        },
        nextEditors(value: readonly CodeEditorData[]): void {
            editors.next(value)
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
