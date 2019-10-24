import { WorkspaceRoot } from '@sourcegraph/extension-api-types'
import { BehaviorSubject, NextObserver, Subscribable } from 'rxjs'

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
 * The workspace service manages the workspace root folders.
 */
export interface WorkspaceService {
    /**
     * An observable whose value is the list of workspace root folder URIs.
     */
    readonly roots: Subscribable<readonly WorkspaceRootWithMetadata[]> & {
        readonly value: readonly WorkspaceRootWithMetadata[]
    } & NextObserver<readonly WorkspaceRootWithMetadata[]>
}

/**
 * Creates a {@link WorkspaceService} instance.
 */
export function createWorkspaceService(): WorkspaceService {
    const roots = new BehaviorSubject<readonly WorkspaceRootWithMetadata[]>([])
    return {
        roots,
    }
}
