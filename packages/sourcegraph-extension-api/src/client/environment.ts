import { ConfigurationCascade } from '../protocol'
import { Context, EMPTY_CONTEXT } from './context/context'
import { Extension } from './extension'
import { TextDocumentItem } from './types/textDocument'

/**
 * A description of the environment represented by the Sourcegraph extension client application.
 *
 * This models the state of editor-like tools that display documents, allow selections and scrolling
 * in documents, and support extension configuration.
 *
 * @template X extension type, to support storing additional properties on extensions
 * @template C configuration cascade type
 */
export interface Environment<X extends Extension = Extension, C extends ConfigurationCascade = ConfigurationCascade> {
    /**
     * The text documents that are currently visible. Each text document is represented to extensions as being
     * in its own visible CodeEditor.
     */
    readonly visibleTextDocuments: TextDocumentItem[] | null

    /** The active extensions, or null if there are none. */
    readonly extensions: X[] | null

    /** The configuration cascade. */
    readonly configuration: C

    /** Arbitrary key-value pairs that describe other application state. */
    readonly context: Context
}

/** An empty Sourcegraph extension client environment. */
export const EMPTY_ENVIRONMENT: Environment<any, any> = {
    visibleTextDocuments: null,
    extensions: null,
    configuration: { merged: {} },
    context: EMPTY_CONTEXT,
}
