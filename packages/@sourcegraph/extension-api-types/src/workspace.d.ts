/**
 * A root directory of a workspace.
 *
 * @see module:sourcegraph.WorkspaceRoot
 */
export interface WorkspaceRoot {
    /** The root URI of the workspace. */
    readonly uri: string

    /** The base URI, used when the workspace is opened for comparison. */
    readonly baseUri?: string
}
