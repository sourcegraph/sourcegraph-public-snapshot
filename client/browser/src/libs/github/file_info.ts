import { isDefined, propertyIsDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { Observable, of, throwError, zip } from 'rxjs'
import { filter, map, switchMap } from 'rxjs/operators'
import { GitHubBlobUrl } from '.'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { getCommitIDFromPermalink } from './scrape'
import { getDeltaFileName, getDiffResolvedRev, getGitHubState, parseURL } from './util'

export const resolveDiffFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    of(codeView).pipe(
        map(codeView => {
            const { repoName } = parseURL()

            return { codeView, repoName }
        }),
        map(({ codeView, ...rest }) => {
            const { headFilePath, baseFilePath } = getDeltaFileName(codeView)
            if (!headFilePath) {
                throw new Error('cannot determine file path')
            }

            return { ...rest, codeView, headFilePath, baseFilePath }
        }),
        map(data => {
            const diffResolvedRev = getDiffResolvedRev(codeView)
            if (!diffResolvedRev) {
                throw new Error('cannot determine delta info')
            }

            return {
                headRev: diffResolvedRev.headCommitID,
                baseRev: diffResolvedRev.baseCommitID,
                ...data,
            }
        }),
        switchMap(({ repoName, headRev, baseRev, ...rest }) => {
            const resolvingHeadRev = resolveRev({ repoName, rev: headRev, requestGraphQL }).pipe(
                retryWhenCloneInProgressError()
            )
            const resolvingBaseRev = resolveRev({ repoName, rev: baseRev, requestGraphQL }).pipe(
                retryWhenCloneInProgressError()
            )

            return zip(resolvingHeadRev, resolvingBaseRev).pipe(
                map(([headCommitID, baseCommitID]) => ({
                    repoName,
                    headRev,
                    baseRev,
                    headCommitID,
                    baseCommitID,
                    ...rest,
                }))
            )
        }),
        map(info => ({
            repoName: info.repoName,
            filePath: info.headFilePath,
            commitID: info.headCommitID,
            rev: info.headRev,

            baseRepoName: info.repoName,
            baseFilePath: info.baseFilePath || info.headFilePath,
            baseCommitID: info.baseCommitID,
            baseRev: info.baseRev,

            headHasFileContents: true,
            baseHasFileContents: true,
        }))
    )

export const resolveFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    const { repoName, filePath, rev } = parseURL()
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
            repoName,
            filePath,
            commitID,
            rev: rev || commitID,
        })
    } catch (error) {
        return throwError(error)
    }
}

export const resolveSnippetFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
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
        filter(propertyIsDefined('ghRepoName')),
        filter(propertyIsDefined('rev')),
        filter(propertyIsDefined('filePath')),
        map(({ owner, ghRepoName, ...rest }) => ({
            repoName: `${window.location.host}/${owner}/${ghRepoName}`,
            ...rest,
        })),
        switchMap(({ repoName, rev, ...rest }) =>
            resolveRev({ repoName, rev, requestGraphQL }).pipe(
                retryWhenCloneInProgressError(),
                map(commitID => ({ ...rest, repoName, commitID, rev: rev || commitID }))
            )
        )
    )
