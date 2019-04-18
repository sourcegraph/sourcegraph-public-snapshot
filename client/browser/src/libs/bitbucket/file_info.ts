import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { FileInfo } from '../code_intelligence'
import { getBaseCommit, getCommitsForPR } from './api'
import {
    getCommitInfoFromComparePage,
    getFileInfoFromCommitDiffCodeView,
    getFileInfoFromSingleFileDiffCodeView,
    getFileInfoFromSingleFileSourceCodeView,
    getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView,
    getPRIDFromPathName,
} from './scrape'

/**
 * Resolves file information for a page with a single file in source (not diff) view.
 */
export const resolveFileInfoForSingleFileSourceView = (codeView: HTMLElement): Observable<FileInfo> => {
    const fileInfo = getFileInfoFromSingleFileSourceCodeView(codeView)
    return of(fileInfo)
}

/**
 * Gets the file info for a PR diff code view.
 */
export const resolvePullRequestFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
    const prID = getPRIDFromPathName()
    return getCommitsForPR({ ...fileInfo, prID }).pipe(
        map(({ headCommitID, baseCommitID }) => ({ ...fileInfo, commitID: headCommitID, baseCommitID }))
    )
}

/**
 * Gets the file info for a single-file "diff to previous" code view.
 */
export const resolveSingleFileDiffFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    const fileInfo = getFileInfoFromSingleFileDiffCodeView(codeView)
    return getBaseCommit(fileInfo).pipe(map(baseCommitID => ({ baseCommitID, ...fileInfo })))
}

export const resolveCommitViewFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(getFileInfoFromCommitDiffCodeView(codeView))

/**
 * Resolves the file info on a compare page.
 */
export const resolveCompareFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(codeView).pipe(
        map(codeView => {
            const { baseCommitID, headCommitID } = getCommitInfoFromComparePage()
            return {
                ...getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView),
                commitID: headCommitID,
                baseCommitID,
            }
        })
    )
