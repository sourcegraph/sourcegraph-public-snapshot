import { Observable, of, throwError, from } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { FileInfo } from '../code_intelligence'

import { getBaseCommitIDForCommit, getMergeRequestDetailsFromAPI } from './api'
import {
    getCommitIDFromPermalink,
    getCommitPageInfo,
    getFilePageInfo,
    getFilePathsFromCodeView,
    getPageInfo,
    getMergeRequestID,
    getDiffID,
} from './scrape'

/**
 * Resolves file information for a page with a single file, not including diffs with only one file.
 */
export const resolveFileInfo = (): Observable<FileInfo> => {
    const { rawRepoName, filePath, rev } = getFilePageInfo()
    if (!filePath) {
        return throwError(
            new Error(
                `Unable to determine the file path of the current file because the current URL (window.location ${window.location}) does not have a file path.`
            )
        )
    }

    try {
        const commitID = getCommitIDFromPermalink()

        return of({
            rawRepoName,
            filePath,
            commitID,
            rev,
        })
    } catch (error) {
        return throwError(error)
    }
}

/**
 * Gets `FileInfo` for a diff file.
 */
export const resolveDiffFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    from(
        getMergeRequestDetailsFromAPI({
            ...getPageInfo(),
            mergeRequestID: getMergeRequestID(),
            diffID: getDiffID(),
        })
    ).pipe(map((info): FileInfo => ({ ...info, ...getFilePathsFromCodeView(codeView) })))

/**
 * Resolves file information for commit pages.
 */
export const resolveCommitFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(undefined).pipe(
        map(getCommitPageInfo),
        // Resolve base commit ID.
        switchMap(({ owner, projectName, commitID, rawRepoName }) =>
            getBaseCommitIDForCommit({ owner, projectName, commitID }).pipe(
                map(baseCommitID => ({ commitID, baseCommitID, rawRepoName }))
            )
        ),
        map(
            ({ commitID, baseCommitID, rawRepoName }): FileInfo => {
                const { filePath, baseFilePath } = getFilePathsFromCodeView(codeView)
                return { baseCommitID, baseFilePath, commitID, filePath, rawRepoName }
            }
        )
    )
