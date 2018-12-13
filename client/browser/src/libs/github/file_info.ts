import { isDefined, propertyIsDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { Observable, of, throwError, zip } from 'rxjs'
import { filter, map, switchMap } from 'rxjs/operators'
import { GitHubBlobUrl } from '.'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { ensureRevisionsAreCloned } from '../code_intelligence/util/file_info'
import { getCommitIDFromPermalink } from './scrape'
import { getDeltaFileName, getDiffResolvedRev, getGitHubState, parseURL } from './util'

export const resolveDiffFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(codeView).pipe(
        map(codeView => {
            const { repoPath } = parseURL()

            return { codeView, repoPath }
        }),
        map(({ codeView, ...rest }) => {
            const { headFilePath, baseFilePath } = getDeltaFileName(codeView)
            if (!headFilePath) {
                throw new Error('cannot determine file path')
            }

            return { ...rest, codeView, headFilePath, baseFilePath }
        }),
        map(data => {
            const diffResolvedRev = getDiffResolvedRev()
            if (!diffResolvedRev) {
                throw new Error('cannot determine delta info')
            }

            return {
                headRev: diffResolvedRev.headCommitID,
                baseRev: diffResolvedRev.baseCommitID,
                ...data,
            }
        }),
        switchMap(({ repoPath, headRev, baseRev, ...rest }) => {
            const resolvingHeadRev = resolveRev({ repoPath, rev: headRev }).pipe(retryWhenCloneInProgressError())
            const resolvingBaseRev = resolveRev({ repoPath, rev: baseRev }).pipe(retryWhenCloneInProgressError())

            return zip(resolvingHeadRev, resolvingBaseRev).pipe(
                map(([headCommitID, baseCommitID]) => ({
                    repoPath,
                    headRev,
                    baseRev,
                    headCommitID,
                    baseCommitID,
                    ...rest,
                }))
            )
        }),
        map(info => ({
            repoPath: info.repoPath,
            filePath: info.headFilePath,
            commitID: info.headCommitID,
            rev: info.headRev,

            baseRepoPath: info.repoPath,
            baseFilePath: info.baseFilePath || info.headFilePath,
            baseCommitID: info.baseCommitID,
            baseRev: info.baseRev,

            headHasFileContents: true,
            baseHasFileContents: true,
        }))
    )

export const resolveFileInfo = (): Observable<FileInfo> => {
    const { repoPath, filePath, rev } = parseURL()
    if (!filePath) {
        return throwError(
            new Error(
                `Unable to determine the file path of the current file because the current URL (window.location ${
                    window.location
                }) does not have a file path.`
            )
        )
    }

    try {
        const commitID = getCommitIDFromPermalink()

        return of({
            repoPath,
            filePath,
            commitID,
            rev: rev || commitID,
        }).pipe(ensureRevisionsAreCloned)
    } catch (error) {
        return throwError(error)
    }
}

export const resolveSnippetFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(codeView).pipe(
        map(codeView => {
            const anchors = codeView.getElementsByTagName('a')
            let githubState: GitHubBlobUrl | undefined
            for (const anchor of anchors) {
                const anchorState = getGitHubState(anchor.href) as GitHubBlobUrl
                if (anchorState) {
                    githubState = anchorState
                    break
                }
            }

            return githubState
        }),
        filter(isDefined),
        filter(propertyIsDefined('owner')),
        filter(propertyIsDefined('repoName')),
        filter(propertyIsDefined('rev')),
        filter(propertyIsDefined('filePath')),
        map(({ owner, repoName, ...rest }) => ({ repoPath: `${window.location.host}/${owner}/${repoName}`, ...rest })),
        switchMap(({ repoPath, rev, ...rest }) =>
            resolveRev({ repoPath, rev }).pipe(
                retryWhenCloneInProgressError(),
                map(commitID => ({ ...rest, repoPath, commitID, rev: rev || commitID }))
            )
        )
    )
