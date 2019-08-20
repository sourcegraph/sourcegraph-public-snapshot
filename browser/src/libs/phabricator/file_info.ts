import { Observable, zip } from 'rxjs'
import { filter, map, switchMap, tap } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { FileInfo } from '../code_intelligence'
import { DifferentialState, DiffusionState, PhabricatorMode, RevisionState } from '.'
import { resolveDiffRev } from './backend'
import { getFilepathFromFileForDiff, getFilePathFromFileForRevision } from './scrape'
import { getPhabricatorState } from './util'

export const resolveRevisionFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    getPhabricatorState(window.location, requestGraphQL).pipe(
        filter((state): state is RevisionState => state.mode === PhabricatorMode.Revision),
        map(({ rawRepoName, headCommitID, baseCommitID }) => ({
            rawRepoName,
            commitID: headCommitID,
            baseCommitID,
        })),
        map(info => ({
            ...info,
            filePath: getFilePathFromFileForRevision(codeView),
        }))
    )

export const resolveDiffFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    getPhabricatorState(window.location, requestGraphQL).pipe(
        filter((state): state is DifferentialState => state.mode === PhabricatorMode.Differential),
        switchMap(state => {
            const { filePath, baseFilePath } = getFilepathFromFileForDiff(codeView)
            const resolveBaseCommitID = resolveDiffRev(
                {
                    repoName: state.baseRawRepoName,
                    revisionID: state.revisionID,
                    diffID: state.baseDiffID || state.diffID,
                    baseDiffID: state.baseDiffID,
                    useDiffForBase: Boolean(state.baseDiffID), // if ?vs and base is not `on` i.e. the initial commit)
                    useBaseForDiff: false,
                    filePath: baseFilePath || filePath,
                    isBase: true,
                },
                requestGraphQL
            ).pipe(
                tap(info => {
                    console.log('resolveBaseCommitID', info)
                }),
                map(({ commitID, stagingRepoName }) => ({
                    baseCommitID: commitID,
                    baseRawRepoName: stagingRepoName || state.baseRawRepoName,
                }))
            )
            const resolveHeadCommitID = resolveDiffRev(
                {
                    repoName: state.headRawRepoName,
                    revisionID: state.revisionID,
                    diffID: state.diffID,
                    baseDiffID: state.baseDiffID,
                    useDiffForBase: false,
                    useBaseForDiff: false,
                    filePath,
                    isBase: false,
                },
                requestGraphQL
            ).pipe(
                tap(info => {
                    console.log('resolveHeadCommitID', info)
                }),
                map(({ commitID, stagingRepoName }) => ({
                    commitID,
                    rawRepoName: stagingRepoName || state.headRawRepoName,
                }))
            )
            return zip(resolveBaseCommitID, resolveHeadCommitID).pipe(
                map(
                    ([{ baseCommitID, baseRawRepoName }, { commitID, rawRepoName }]): FileInfo => ({
                        ...state,
                        baseCommitID,
                        commitID,
                        filePath,
                        baseFilePath: baseFilePath || filePath,
                        baseRawRepoName,
                        rawRepoName,
                    })
                )
            )
        })
    )

export const resolveDiffusionFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    getPhabricatorState(window.location, requestGraphQL).pipe(
        filter((state): state is DiffusionState => state.mode === PhabricatorMode.Diffusion)
    )
