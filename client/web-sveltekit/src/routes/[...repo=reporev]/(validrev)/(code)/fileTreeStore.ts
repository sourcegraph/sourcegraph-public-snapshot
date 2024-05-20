import { from, of, Subject } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { readable, type Readable } from 'svelte/store'

import { browser } from '$app/environment'
import { FileTreeProvider, ROOT_PATH, type FileTreeData, type FileTreeLoader } from '$lib/repo/api/tree'

/**
 * Keeps track of the top-level directory that has been visited for each repository and revision.
 */
const topTreePathByRepoAndRevision = new Map<string, Map<string, string>>()

/**
 * Clears the cache of top-level directories that have been visited.
 * This should only be used in tests.
 */
export function clearTopTreePathCache_testingOnly(): void {
    topTreePathByRepoAndRevision.clear()
}

/**
 * Manages the state of the sidebar file tree.
 *
 * @remarks
 * This store ensures that we always show the most top-level directory that has been visited so far.
 */
interface FileTreeStore extends Readable<FileTreeProvider | Error | null> {
    /**
     * Sets the current repo, revision, and path for the file tree.
     */
    set(args: { repoName: string; revision: string; path: string }): void

    /**
     * Resets file tree top level path, in some cases like jump to scope
     * we have to invalidate cache since top level path has been changed
     * to a lower level.
     */
    resetTopPathCache(repoName: string, revision: string): void
}

interface FileTreeStoreOptions {
    /**
     * Fetches the file tree for the given repo, revision, and path.
     */
    fetchFileTreeData: (args: { repoName: string; revision: string; filePath: string }) => Promise<FileTreeData>
}

/**
 * Helper function for managing the sidebar file tree state.
 * Specifically it ensures that we always show the most top-level directory
 * that has been visited so far.
 */
export function createFileTreeStore(options: FileTreeStoreOptions): FileTreeStore {
    const repoRevPath = new Subject<{ repoName: string; revision: string; path: string }>()
    const { subscribe } = readable<FileTreeProvider | Error | null>(null, set => {
        const subscription = repoRevPath
            .pipe(
                // We need to create a new file tree provider in the following cases:
                // - The repo changes
                // - The revision changes
                // - The path is not a subdirectory of the top path
                distinctUntilChanged(
                    ({ repoName, revision }, { repoName: nextRepoName, revision: nextRevision, path: nextPath }) => {
                        if (browser && repoName === nextRepoName && revision === nextRevision) {
                            const topPath = topTreePathByRepoAndRevision.get(repoName)?.get(revision)
                            return topPath !== undefined ? topPath === ROOT_PATH || nextPath.startsWith(topPath) : false
                        }
                        return false
                    }
                ),
                // If the path is not a subdirectory of the top path, we need to update the top path, otherwise we use the top path
                map(({ repoName, revision, path }) => {
                    if (browser) {
                        const topPath = topTreePathByRepoAndRevision.get(repoName)?.get(revision)
                        if (topPath !== undefined && (topPath === ROOT_PATH || path.startsWith(topPath))) {
                            return { repoName, revision, path: topPath }
                        } else {
                            // new path is new top path
                            const topPaths = topTreePathByRepoAndRevision.get(repoName) || new Map()
                            topPaths.set(revision, path)
                            topTreePathByRepoAndRevision.set(repoName, topPaths)
                        }
                    }
                    return { repoName, revision, path }
                }),
                // Fetch the file tree for the given repo, revision, and path
                switchMap(({ repoName, revision, path }) => {
                    const loader: FileTreeLoader = args =>
                        options.fetchFileTreeData({ repoName, revision, filePath: args.filePath })
                    return from(loader({ filePath: path })).pipe(
                        map(data => new FileTreeProvider({ ...data, loader })),
                        // If an observable errors the subscription is closed, so we need to catch
                        // the error here to ensure that the (outer) subscription stays open
                        catchError(error => of(error))
                    )
                })
            )
            .subscribe(set)
        return () => {
            subscription.unsubscribe()
        }
    })

    return {
        subscribe,
        set(args) {
            repoRevPath.next(args)
        },
        resetTopPathCache(repoName: string, revision: string): void {
            const topPaths = topTreePathByRepoAndRevision.get(repoName) || new Map()
            topPaths.set(revision, undefined)
        },
    }
}
