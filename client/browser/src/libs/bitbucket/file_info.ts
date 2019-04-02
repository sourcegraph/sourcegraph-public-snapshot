import { propertyIsDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { Observable, of } from 'rxjs'
import { filter, map, switchMap } from 'rxjs/operators'

import { DiffResolvedRevSpec } from '../../shared/repo'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { ensureRevisionsAreCloned } from '../code_intelligence/util/file_info'
import { getBaseCommit, getCommitsForPR } from './api'
import {
    getDiffFileInfoFromCodeView,
    getFileInfoFromCodeView,
    getPRIDFromPathName,
    getResolvedDiffFromBranchComparePage,
} from './scrape'

/**
 * Resolves file information for a page with a single file, not including diffs with only one file.
 */
export const resolveFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(codeView).pipe(
        map(getFileInfoFromCodeView),
        filter(propertyIsDefined('filePath')),
        ensureRevisionsAreCloned
    )

const fetchDiffFiles = (
    info: Pick<FileInfo, 'repoName' | 'filePath' | 'rev'> & DiffResolvedRevSpec
): Observable<FileInfo> =>
    of(info).pipe(
        map(({ headCommitID, ...rest }) => ({ ...rest, commitID: headCommitID })),
        switchMap(info =>
            fetchBlobContentLines(info).pipe(
                map(lines => ({
                    ...info,
                    content: lines.join('\n'),
                    headHasFileContents: lines.length > 0,
                }))
            )
        ),
        switchMap(({ repoName, filePath, baseCommitID, ...rest }) =>
            fetchBlobContentLines({
                repoName,
                filePath,
                commitID: baseCommitID,
            }).pipe(
                map(lines => ({
                    repoName,
                    filePath,
                    baseCommitID,
                    baseContent: lines.join('\n'),
                    baseHasFileContents: lines.length > 0,
                    ...rest,
                }))
            )
        )
    )

export const resolveDiffFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(codeView).pipe(
        map(getDiffFileInfoFromCodeView),
        switchMap(({ commitID, project, repoSlug, ...rest }) => {
            if (commitID) {
                return getBaseCommit({ commitID, project, repoSlug }).pipe(
                    map(baseCommitID => ({ baseCommitID, headCommitID: commitID, ...rest }))
                )
            }
            const prID = getPRIDFromPathName()
            return getCommitsForPR({ project, repoSlug, prID }).pipe(map(commits => ({ ...rest, ...commits })))
        }),
        switchMap(fetchDiffFiles)
    )

export const resolveCompareFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    of(codeView).pipe(
        map(codeView => ({
            ...getDiffFileInfoFromCodeView(codeView),
            ...getResolvedDiffFromBranchComparePage(),
        })),
        switchMap(fetchDiffFiles)
    )
