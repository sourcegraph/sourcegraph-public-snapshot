import { WorkspaceRoot } from '../protocol/plainTypes'
import { TextDocumentItem } from './types/textDocument'

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
     * The text documents that are currently visible. Each text document is represented to extensions as being
     * in its own visible CodeEditor.
     */
    readonly visibleTextDocuments: TextDocumentItem[] | null
}

/** An empty Sourcegraph extension client model. */
export const EMPTY_MODEL: Model = {
    roots: null,
    visibleTextDocuments: null,
}
