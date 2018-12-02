import { Selection, WorkspaceRoot } from '../protocol/plainTypes'
import { TextDocumentItem } from './types/textDocument'

/**
 * Describes a view component.
 *
 * @todo Currently the only view component is CodeEditor ("textEditor" as exposed in the API), so this type just
 * describes a CodeEditor. When more view components exist, this type will need to become a union type or add in
 * some other similar abstraction to support describing all types of view components.
 */
export interface ViewComponentData {
    type: 'textEditor'
    item: TextDocumentItem
    selections: Selection[]
}

/**
 * A description of the model represented by the Sourcegraph extension client application.
 *
 * This models the state of editor-like tools that display documents, allow selections and scrolling
 * in documents, and support extension configuration.
 */
export interface Model {
    /**
     * The currently open workspace roots (typically a single repository).
     */
    readonly roots: WorkspaceRoot[] | null

    /**
     * The view components that are currently visible. Each text document is represented as being in its own
     * visible CodeEditor.
     */
    readonly visibleViewComponents: ViewComponentData[] | null
}

/** An empty Sourcegraph extension client model. */
export const EMPTY_MODEL: Model = {
    roots: null,
    visibleViewComponents: null,
}
