import { Observable, of, throwError, zip } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { getCommitIDFromPermalink } from './scrape'
import { getBranchName, getDeltaFileName, getDiffResolvedRev, parseURL } from './util'

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
    try {
        const parsedURL = parseURL()
        if (parsedURL.pageType !== 'blob') {
            throw new Error(`Current URL does not match a blob url: ${window.location}`)
        }
        const { revAndFilePath, repoName } = parsedURL

        // We must scrape for the branch name in order to determine
        // the precise file path: otherwise we'll get tripped up parsing the
        // pathname when the branch name contains forward slashes.
        // TODO ideally, this should only scrape the code view itself.
        const branchName = getBranchName()
        if (!revAndFilePath.startsWith(branchName)) {
            throw new Error(
                `Could not parse filePath: revAndFilePath ${revAndFilePath} does not start with branchName ${branchName}`
            )
        }
        const filePath = revAndFilePath.slice(branchName.length + 1)
        const commitID = getCommitIDFromPermalink()
        return of({
            repoName,
            filePath,
            commitID,
            rev: branchName,
        })
    } catch (error) {
        return throwError(error)
    }
}

const COMMIT_HASH_REGEX = /\/([0-9a-f]{40})$/i

export const resolveSnippetFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    try {
        // A snippet code view contains a link to the snippet's commit.
        // We use it to find the 40-character commit id.
        const commitLinkElement = codeView.querySelector('a.commit-tease-sha') as HTMLAnchorElement
        if (!commitLinkElement) {
            throw new Error('Could not find commit link in snippet code view')
        }
        const commitIDMatch = commitLinkElement.href.match(COMMIT_HASH_REGEX)
        if (!commitIDMatch || !commitIDMatch[1]) {
            throw new Error(`Could not parse commitID from snippet commit link href: ${commitLinkElement.href}`)
        }
        const commitID = commitIDMatch[1]

        // We then use the permalink to determine the repo name and parse the filePath.
        const selector = 'a:not(.commit-tease-sha)'
        const anchors = codeView.querySelectorAll(selector)
        const snippetPermalinkURL = new URL((anchors[0] as HTMLAnchorElement).href)
        const parsedURL = parseURL(snippetPermalinkURL)
        if (parsedURL.pageType !== 'blob') {
            throw new Error(`Snippet URL does not match a blob url: ${snippetPermalinkURL}`)
        }
        const { revAndFilePath, repoName } = parsedURL
        if (!revAndFilePath.startsWith(commitID)) {
            throw new Error(
                `Could not parse filePath: revAndFilePath ${revAndFilePath} does not start with commitID ${commitID}`
            )
        }
        const filePath = revAndFilePath.slice(commitID.length + 1)
        return of({
            repoName,
            filePath,
            commitID,
            rev: commitID,
        })
    } catch (err) {
        return throwError(err)
    }
}
