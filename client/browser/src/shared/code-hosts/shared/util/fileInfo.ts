import { from, type Observable, of, zip } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'

import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { resolveRepo, resolveRevision, retryWhenCloneInProgressError } from '../../../repo/backend'
import type { FileInfo, FileInfoWithRepoName, DiffOrBlobInfo } from '../codeHost'

/**
 * Use `rawRepoName` for the value of `repoName`, as a fallback if `repoName`
 * was not available.
 */
const useRawRepoNameAsFallback = (fileInfo: FileInfo): FileInfoWithRepoName => ({
    ...fileInfo,
    repoName: fileInfo.rawRepoName,
})

/**
 * Resolve a `FileInfo`'s raw repo names to their Sourcegraph
 * repo names as affected by `repositoryPathPattern`.
 */
const resolveRepoNameForFileInfo = (
    fileInfo: FileInfo,
    checkRepoSyncError: (error: any) => Promise<boolean>,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfoWithRepoName> =>
    resolveRepo({ rawRepoName: fileInfo.rawRepoName, requestGraphQL }).pipe(
        map(repoName => ({ ...fileInfo, repoName })),
        catchError(error =>
            // Check if the repository either is a private repository
            // that has not been found (if the browser extension is pointed towards
            // Sourcegraph Cloud) or is not added to the Sourcegraph instance
            // (if the active Sourcegraph instance is other than Cloud).
            // In that case, it's impossible to resolve the repo names,
            // so we keep the repo names inferred from the code host's DOM.
            // Note: we recover/fallback in this case so that we can show informative
            // alerts to the user.
            from(checkRepoSyncError(error)).pipe(
                switchMap(hasRepoSyncError => {
                    if (hasRepoSyncError) {
                        return of(useRawRepoNameAsFallback(fileInfo))
                    }
                    throw error
                })
            )
        )
    )

/**
 * Uses the `commitID` as the default value for `revision`, if no `revision`
 * value is present.
 */
export const defaultRevisionToCommitID = <T extends FileInfo>(fileInfo: T): T & { revision: string } => ({
    ...fileInfo,
    revision: fileInfo.revision || fileInfo.commitID,
})

/**
 * Ensures that the revision is cloned on Sourcegraph, by issuing a
 * `resolveRevision` request and retrying until the clone is complete.
 */
export const ensureRevisionIsClonedForFileInfo = (
    fileInfo: FileInfoWithRepoName,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<string> => {
    // Although we get the commit SHA's from elsewhere, we still need to
    // use `resolveRev` otherwise we can't guarantee Sourcegraph has the
    // revision cloned.
    const { repoName, commitID } = fileInfo
    return resolveRevision({ repoName, revision: commitID, requestGraphQL }).pipe(retryWhenCloneInProgressError())
}

export const resolveRepoNamesForDiffOrFileInfo = (
    diffOrFileInfo: DiffOrBlobInfo,
    checkRepoSyncError: (error: any) => Promise<boolean>,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<DiffOrBlobInfo<FileInfoWithRepoName>> => {
    if ('blob' in diffOrFileInfo) {
        return resolveRepoNameForFileInfo(diffOrFileInfo.blob, checkRepoSyncError, requestGraphQL).pipe(
            map(fileInfo => ({ blob: { ...diffOrFileInfo.blob, ...fileInfo } }))
        )
    }
    if (diffOrFileInfo.head && diffOrFileInfo.base) {
        const resolvingHeadWithRepoName = resolveRepoNameForFileInfo(
            diffOrFileInfo.head,
            checkRepoSyncError,
            requestGraphQL
        )
        const resolvingBaseWithRepoName = resolveRepoNameForFileInfo(
            diffOrFileInfo.base,
            checkRepoSyncError,
            requestGraphQL
        )

        return zip(resolvingHeadWithRepoName, resolvingBaseWithRepoName).pipe(map(([head, base]) => ({ head, base })))
    }
    if (diffOrFileInfo.head) {
        return resolveRepoNameForFileInfo(diffOrFileInfo.head, checkRepoSyncError, requestGraphQL).pipe(
            map(head => ({ head }))
        )
    }
    return resolveRepoNameForFileInfo(diffOrFileInfo.base, checkRepoSyncError, requestGraphQL).pipe(
        map(base => ({ base }))
    )
}
