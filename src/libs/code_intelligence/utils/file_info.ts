import { Observable, zip } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { FileInfo } from '..'
import { resolveRev, retryWhenCloneInProgressError } from '../../../shared/repo/backend'

export const ensureRevisionsAreCloned = (files: Observable<FileInfo>): Observable<FileInfo> =>
    files.pipe(
        switchMap(({ repoPath, rev, baseRev, ...rest }) => {
            // Although we get the commit SHA's from elesewhere, we still need to
            // use `resolveRev` otherwise we can't guarantee Sourcegraph has the
            // revision cloned.
            const resolvingHeadRev = resolveRev({ repoPath, rev }).pipe(retryWhenCloneInProgressError())
            const resolvingBaseRev = resolveRev({ repoPath, rev: baseRev }).pipe(retryWhenCloneInProgressError())

            return zip(resolvingHeadRev, resolvingBaseRev).pipe(map(() => ({ repoPath, rev, baseRev, ...rest })))
        })
    )
