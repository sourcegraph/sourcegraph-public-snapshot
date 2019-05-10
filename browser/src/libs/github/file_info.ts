import { Observable, of, throwError, zip } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { getCommitIDFromPermalink } from './scrape'
import { getDeltaFileName, getDiffResolvedRev, parseURL } from './util'

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

const COMMIT_HASH_REGEX = /^[0-9a-f]{40}$/i

export const resolveSnippetFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    try {
        const selector = 'a:not(.commit-tease-sha)'
        const anchors = codeView.querySelectorAll(selector)
        const snippetPermalinkURL = new URL((anchors[0] as HTMLAnchorElement).href)
        if (anchors.length !== 1) {
            throw new Error(`Found ${anchors.length} matching ${selector} in snippet code view`)
        }
        const { repoName, filePath, rev } = parseURL(snippetPermalinkURL)
        if (!filePath || !rev || !rev.match(COMMIT_HASH_REGEX)) {
            throw new Error(`Could not determine snippet FileInfo from permalink href ${snippetPermalinkURL}`)
        }
        return of({
            repoName,
            filePath,
            // The rev in a snippet permalink is a 40-character commit sha
            commitID: rev,
            rev,
        })
    } catch (err) {
        return throwError(err)
    }
}
