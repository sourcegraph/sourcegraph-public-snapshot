import { Observable, of, zip } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { isPrivateRepoPublicSourcegraphComErrorLike } from '../../../../../../shared/src/backend/errors'
import { PlatformContext } from '../../../../../../shared/src/platform/context'
import { resolveRepo, resolveRev, retryWhenCloneInProgressError } from '../../../repo/backend'
import { FileInfo, FileInfoWithRepoName, DiffOrBlobInfo } from '../codeHost'

export const ensureRevisionIsClonedForFileInfo = (
    fileInfo: FileInfoWithRepoName,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<string> => {
    // Although we get the commit SHA's from elsewhere, we still need to
    // use `resolveRev` otherwise we can't guarantee Sourcegraph has the
    // revision cloned.
    const { repoName, commitID } = fileInfo
    return resolveRev({ repoName, rev: commitID, requestGraphQL }).pipe(retryWhenCloneInProgressError())
}

export const resolveRepoNamesForDiffOrFileInfo = (
    diffOrFileInfo: DiffOrBlobInfo,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<DiffOrBlobInfo<FileInfoWithRepoName>> => {
    if ('blob' in diffOrFileInfo) {
        return resolveRepoNameForFileInfo(diffOrFileInfo.blob, requestGraphQL).pipe(
            map(fileInfo => ({ ...diffOrFileInfo, blob: { ...diffOrFileInfo.blob, ...fileInfo } }))
        )
    } else if ('head' in diffOrFileInfo && 'base' in diffOrFileInfo) {
        const resolvingHeadWithRepoName = resolveRepoNameForFileInfo(diffOrFileInfo.head, requestGraphQL)
        const resolvingBaseWithRepoName = resolveRepoNameForFileInfo(diffOrFileInfo.base, requestGraphQL)

        return zip(resolvingHeadWithRepoName, resolvingBaseWithRepoName).pipe(
            map(([head, base]) => ({
                ...diffOrFileInfo,
                head,
                base,
            }))
        )
    } else if ('head' in diffOrFileInfo) {
        return resolveRepoNameForFileInfo(diffOrFileInfo.head, requestGraphQL).pipe(
            map(head => ({
                ...diffOrFileInfo,
                head,
            }))
        )
    } else if ('base' in diffOrFileInfo) {
        return resolveRepoNameForFileInfo(diffOrFileInfo.base, requestGraphQL).pipe(
            map(base => ({
                ...diffOrFileInfo,
                base,
            }))
        )
    }
    throw new Error('Cannot resolve file info: must contain a blob, base, or head.')
}

/**
 * Use `rawRepoName` for the value of `repoName`, as a fallback if `repoName` was not available.
 * */
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
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfoWithRepoName> =>
    resolveRepo({ rawRepoName: fileInfo.rawRepoName, requestGraphQL }).pipe(
        map(repoName => ({ ...fileInfo, repoName })),
        catchError(err => {
            // ERPRIVATEREPOPUBLICSOURCEGRAPHCOM likely means that the user is viewing private code
            // without having pointed his browser extension to a self-hosted Sourcegraph instance that
            // has access to that code. In that case, it's impossible to resolve the repo names,
            // so we keep the repo names inferred from the code host's DOM.
            if (isPrivateRepoPublicSourcegraphComErrorLike(err)) {
                return of(useRawRepoNameAsFallback(fileInfo))
            }
            throw err
        })
    )

export const ensureRev = <T extends FileInfo>(fileInfo: T): T & { rev: string } => ({
    ...fileInfo,
    rev: fileInfo.rev || fileInfo.commitID,
})
