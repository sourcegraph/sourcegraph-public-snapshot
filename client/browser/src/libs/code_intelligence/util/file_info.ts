import { Observable, zip } from 'rxjs'
import { map } from 'rxjs/operators'

import { resolveRev, retryWhenCloneInProgressError } from '../../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'

export const ensureRevisionsAreCloned = ({
    repoName,
    commitID,
    baseCommitID,
    ...rest
}: FileInfo): Observable<FileInfo> => {
    // Although we get the commit SHA's from elsewhere, we still need to
    // use `resolveRev` otherwise we can't guarantee Sourcegraph has the
    // revision cloned.

    // Head
    const resolvingHeadRev = resolveRev({ repoName, rev: commitID }).pipe(retryWhenCloneInProgressError())

    const requests = [resolvingHeadRev]

    // If theres a base, resolve it as well.
    if (baseCommitID) {
        const resolvingBaseRev = resolveRev({ repoName, rev: baseCommitID }).pipe(retryWhenCloneInProgressError())

        requests.push(resolvingBaseRev)
    }

    return zip(...requests).pipe(map(() => ({ repoName, commitID, baseCommitID, ...rest })))
}
