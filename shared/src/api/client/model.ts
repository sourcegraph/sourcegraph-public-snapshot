import { Selection, WorkspaceRoot } from '@sourcegraph/extension-api-types'
import { TextDocument } from 'sourcegraph'
import { TextDocumentPositionParams } from '../protocol'

/**
 * Describes a view component.
 *
 * @todo Currently the only view component is CodeEditor ("CodeEditor" as exposed in the API), so this type just
 * describes a CodeEditor. When more view components exist, this type will need to become a union type or add in
 * some other similar abstraction to support describing all types of view components.
 */
export interface ViewComponentData {
    type: 'CodeEditor'
    item: TextDocument
    selections: Selection[]
    isActive: boolean
}

/**
 * A workspace root with additional metadata that is not exposed to extensions.
 */
export interface WorkspaceRootWithMetadata extends WorkspaceRoot {
    /**
     * The original input Git revision that the user requested. The {@link WorkspaceRoot#uri} value will contain
     * the Git commit SHA resolved from the input revision, but it is useful to also know the original revision
     * (e.g., to construct URLs for the user that don't result in them navigating from a branch view to a commit
     * SHA view).
     *
     * For example, if the user is viewing the web page https://github.com/alice/myrepo/blob/master/foo.js (note
     * that the URL contains a Git revision "master"), the input revision is "master".
     *
     * The empty string is a valid value (meaning that the default should be used, such as "HEAD" in Git) and is
     * distinct from undefined. If undefined, the Git commit SHA from {@link WorkspaceRoot#uri} should be used.
     */
    inputRevision?: string
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
    readonly roots: WorkspaceRootWithMetadata[] | null

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
