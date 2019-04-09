import { Selection } from '@sourcegraph/extension-api-types'
import { TextDocument } from 'sourcegraph'
import { TextDocumentPositionParams } from '../protocol'

/**
 * Describes all possible view components.
 *
 * @template D The type of text documents referred to by this data. If the document text is managed
 * out-of-band, this can just be an object containing the document URI.
 */
export type ViewComponentData<D extends Pick<TextDocument, 'uri'> = TextDocument> = CodeEditorViewComponentData<D>

/**
 * Describes a code editor view component.
 */
export interface CodeEditorViewComponentData<D extends Pick<TextDocument, 'uri'> = TextDocument> {
    type: 'CodeEditor'
    item: D
    selections: Selection[]
    isActive: boolean
}

/**
 * A description of the model represented by the Sourcegraph extension client application.
 *
 * This models the state of editor-like tools that display documents, allow selections and scrolling
 * in documents, and support extension configuration.
 */
export interface Model {
    /**
     * The view components that are currently visible. Each text document is represented as being in its own
     * visible CodeEditor.
     */
    readonly visibleViewComponents: ViewComponentData[] | null
}

/** An empty Sourcegraph extension client model. */
export const EMPTY_MODEL: Model = {
    visibleViewComponents: null,
}

/**
 * Helper function to converts from {@link Model} to {@link TextDocumentPositionParams}. If the model doesn't have
 * a position, it returns null.
 */
export function modelToTextDocumentPositionParams({
    visibleViewComponents,
}: Pick<Model, 'visibleViewComponents'>): (TextDocumentPositionParams & { textDocument: TextDocument }) | null {
    if (!visibleViewComponents) {
        return null
    }
    const activeViewComponent = visibleViewComponents.find(({ isActive }) => isActive)
    if (!activeViewComponent) {
        return null
    }
    const sel = activeViewComponent.selections[0]
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
        textDocument: activeViewComponent.item,
        position: sel.start,
    }
}
