import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { DiffOrBlobInfo } from '../shared/codeHost'
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
export const resolveFileInfoForSingleFileSourceView = (codeView: HTMLElement): DiffOrBlobInfo => {
    const fileInfo = getFileInfoFromSingleFileSourceCodeView(codeView)
    return { blob: fileInfo }
}

/**
 * Gets the file info for a PR diff code view.
 */
export const resolvePullRequestFileInfo = (codeView: HTMLElement): Observable<DiffOrBlobInfo> => {
    const partialFileInfo = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
    const { rawRepoName, filePath, baseFilePath } = partialFileInfo
    const prID = getPRIDFromPathName()
    return getCommitsForPR({ ...partialFileInfo, prID }).pipe(
        map(
            ({ headCommitID, baseCommitID }): DiffOrBlobInfo => ({
                head: {
                    rawRepoName,
                    filePath,
                    commitID: headCommitID,
                },
                base: {
                    rawRepoName,
                    filePath: baseFilePath,
                    commitID: baseCommitID,
                },
            })
        )
    )
}

/**
 * Gets the file info for a single-file "diff to previous" code view.
 */
export const resolveSingleFileDiffFileInfo = (codeView: HTMLElement): Observable<DiffOrBlobInfo> => {
    const {
        changeType,
        rawRepoName,
        filePath,
        commitID,
        baseFilePath,
        rev,
        ...bitbucketInfo
    } = getFileInfoFromSingleFileDiffCodeView(codeView)
    // TODO: this could be further refactored to skip getBaseCommit if we don't need the baseCommitID.
    return getBaseCommit({ commitID, ...bitbucketInfo }).pipe(
        map(baseCommitID => {
            switch (changeType) {
                case 'ADD':
                    return { head: { rawRepoName, filePath, commitID } }
                case 'DELETE':
                    return { base: { rawRepoName, filePath: baseFilePath, commitID: baseCommitID } }

                case 'RENAME':
                case 'MOVE':
                case 'COPY':
                case 'MODIFY':
                    return {
                        base: { rawRepoName, filePath: baseFilePath, commitID: baseCommitID },
                        head: { rawRepoName, filePath, commitID },
                    }
            }
        })
    )
}

export const resolveCommitViewFileInfo = (codeView: HTMLElement): DiffOrBlobInfo =>
    getFileInfoFromCommitDiffCodeView(codeView)

/**
 * Resolves the file info on a compare page.
 */
export const resolveCompareFileInfo = (codeView: HTMLElement): DiffOrBlobInfo => {
    const { rawRepoName, filePath, baseFilePath } = getFileInfoWithoutCommitIDsFromMultiFileDiffCodeView(codeView)
    const { baseCommitID, headCommitID } = getCommitInfoFromComparePage()
    return {
        base: {
            rawRepoName,
            filePath: baseFilePath,
            commitID: headCommitID,
        },
        head: { rawRepoName, filePath, commitID: baseCommitID },
    }
}
