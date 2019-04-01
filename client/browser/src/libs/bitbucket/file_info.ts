import { forkJoin, Observable, of } from 'rxjs'
import { concatMap, map } from 'rxjs/operators'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { ensureRevisionsAreCloned } from '../code_intelligence/util/file_info'
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
    return of(fileInfo).pipe(ensureRevisionsAreCloned)
}

/**
 * Fetches the contents for the given head and base revisions of a file.
 *
 * @returns The FileInfo including contents
 */
const fetchDiffFiles = (
    fileInfo: Pick<FileInfo, 'repoName' | 'commitID' | 'filePath' | 'baseRepoName' | 'baseCommitID' | 'baseFilePath'>
): Observable<FileInfo> =>
    forkJoin(
        // Head
        fetchBlobContentLines({
            repoName: fileInfo.repoName,
            commitID: fileInfo.commitID,
            filePath: fileInfo.filePath,
        }).pipe(
            map(lines => ({
                content: lines.join('\n'),
                headHasFileContents: lines.length > 0,
            }))
        ),
        // Base (if exists)
        fileInfo.baseFilePath && fileInfo.baseRepoName && fileInfo.baseCommitID
            ? fetchBlobContentLines({
                  repoName: fileInfo.baseRepoName,
                  commitID: fileInfo.baseCommitID,
                  filePath: fileInfo.baseFilePath,
              }).pipe(
                  map(lines => ({
                      baseContent: lines.join('\n'),
                      baseHasFileContents: lines.length > 0,
                  }))
              )
            : of<{ baseContent?: string; baseHasFileContents?: boolean }>({})
    ).pipe(map(([headContents, baseContents]) => ({ ...fileInfo, ...headContents, ...baseContents })))

/**
 * Gets the file info for a PR diff code view.
 */
export const resolvePullRequestFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    const fileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
    const prID = getPRIDFromPathName()
    return getCommitsForPR({ ...fileInfo, prID }).pipe(
        map(({ headCommitID, baseCommitID }) => ({ ...fileInfo, commitID: headCommitID, baseCommitID })),
        concatMap(fetchDiffFiles)
    )
}

/**
 * Gets the file info for a single-file "diff to previous" code view.
 */
export const resolveSingleFileDiffFileInfo = (codeView: HTMLElement): Observable<FileInfo> => {
    const fileInfo = getFileInfoFromSingleFileDiffCodeView(codeView)
    return getBaseCommit(fileInfo).pipe(
        map(baseCommitID => ({ baseCommitID, ...fileInfo })),
        concatMap(fetchDiffFiles)
    )
}

export const resolveCommitViewFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    fetchDiffFiles(getFileInfoFromCommitDiffCodeView(codeView))

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
        }),
        concatMap(fetchDiffFiles)
    )
