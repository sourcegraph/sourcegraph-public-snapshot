import { Observable, of, zip } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { isPrivateRepoPublicSourcegraphComErrorLike } from '../../../../../../shared/src/backend/errors'
import { PlatformContext } from '../../../../../../shared/src/platform/context'
import { resolveRepo, resolveRev, retryWhenCloneInProgressError } from '../../../repo/backend'
import { FileInfo, FileInfoWithRepoNames } from '../codeHost'

export const ensureRevisionsAreCloned = (
    { repoName, commitID, baseCommitID, ...rest }: FileInfoWithRepoNames,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfoWithRepoNames> => {
    // Although we get the commit SHA's from elsewhere, we still need to
    // use `resolveRev` otherwise we can't guarantee Sourcegraph has the
    // revision cloned.

    // Head
    const resolvingHeadRev = resolveRev({ repoName, rev: commitID, requestGraphQL }).pipe(
        retryWhenCloneInProgressError()
    )

    const requests = [resolvingHeadRev]

    // If theres a base, resolve it as well.
    if (baseCommitID) {
        const resolvingBaseRev = resolveRev({ repoName, rev: baseCommitID, requestGraphQL }).pipe(
            retryWhenCloneInProgressError()
        )
        requests.push(resolvingBaseRev)
    }

    return zip(...requests).pipe(map(() => ({ repoName, commitID, baseCommitID, ...rest })))
}

/**
 * Resolve a `FileInfo`'s raw repo names to their Sourcegraph
 * repo names as affected by `repositoryPathPattern`.
 */
export const resolveRepoNames = (
    { rawRepoName, baseRawRepoName, ...rest }: FileInfo,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfoWithRepoNames> => {
    const resolvingHeadRepoName = resolveRepo({ rawRepoName, requestGraphQL }).pipe(retryWhenCloneInProgressError())
    const resolvingBaseRepoName = baseRawRepoName
        ? resolveRepo({ rawRepoName: baseRawRepoName, requestGraphQL }).pipe(retryWhenCloneInProgressError())
        : of(undefined)

    return zip(resolvingHeadRepoName, resolvingBaseRepoName).pipe(
        map(([repoName, baseRepoName]) => ({ repoName, baseRepoName, rawRepoName, baseRawRepoName, ...rest })),

        // ERPRIVATEREPOPUBLICSOURCEGRAPHCOM likely means that the user is viewing private code
        // without having pointed his browser extension to a self-hosted Sourcegraph instance that
        // has access to that code. In that case, it's impossible to resolve the repo names,
        // so we keep the repo names inferred from the code host's DOM.
        catchError(err => {
            if (isPrivateRepoPublicSourcegraphComErrorLike(err)) {
                return [{ rawRepoName, baseRawRepoName, repoName: rawRepoName, baseRepoName: baseRawRepoName, ...rest }]
            }
            throw err
        })
    )
}
