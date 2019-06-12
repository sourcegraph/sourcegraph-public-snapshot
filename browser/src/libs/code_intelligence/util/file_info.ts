import { Observable, of, zip } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { PlatformContext } from '../../../../../shared/src/platform/context'
import { ERPRIVATEREPOPUBLICSOURCEGRAPHCOM } from '../../../shared/backend/errors'
import { resolveRepo, resolveRev, retryWhenCloneInProgressError } from '../../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'

export const ensureRevisionsAreCloned = (
    { repoName, commitID, baseCommitID, ...rest }: FileInfo,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> => {
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
 * Resolve a `FileInfo`'s head and base repo names to their Sourcegraph
 * repo names as affected by `repositoryPathPattern`.
 */
export const resolveRepoNames = (
    { repoName, baseRepoName, ...rest }: FileInfo,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> => {
    const resolvingHeadRepoName = resolveRepo({ repoName, requestGraphQL }).pipe(retryWhenCloneInProgressError())
    const resolvingBaseRepoName = baseRepoName
        ? resolveRepo({ repoName, requestGraphQL }).pipe(retryWhenCloneInProgressError())
        : of(undefined)

    return zip(resolvingHeadRepoName, resolvingBaseRepoName).pipe(
        map(([headRepoName, baseRepoName]) => ({ repoName: headRepoName, baseRepoName, ...rest })),

        // ERPRIVATEREPOPUBLICSOURCEGRAPHCOM likely means that the user is viewing private code
        // without having pointed his browser extension to a self-hosted Sourcegraph instance that
        // has access to that code. In that case, it's impossible to resolve the repo names,
        // so we keep the repo names inferred from the code host's DOM.
        catchError(err => {
            if (err.name === ERPRIVATEREPOPUBLICSOURCEGRAPHCOM) {
                return [{ repoName, baseRepoName, ...rest }]
            }
            throw err
        })
    )
}
