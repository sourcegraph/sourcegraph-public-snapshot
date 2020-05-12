/**
 * A root directory of a workspace.
 *
 * @see module:sourcegraph.WorkspaceRoot
 */
export interface WorkspaceRoot {
    /** The root URI of the workspace. */
    readonly uri: string
    /** The version context of the workspace. */
    versionContext?: string
}
