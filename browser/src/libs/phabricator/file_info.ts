import { from, Observable, zip } from 'rxjs'
import { catchError, filter, map, switchMap } from 'rxjs/operators'
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
    from(getPhabricatorState(window.location, requestGraphQL)).pipe(
        filter((state): state is RevisionState => state !== null && state.mode === PhabricatorMode.Revision),
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
    from(getPhabricatorState(window.location, requestGraphQL)).pipe(
        filter((state): state is DifferentialState => state !== null && state.mode === PhabricatorMode.Differential),
        switchMap(state => {
            const { filePath, baseFilePath } = getFilepathFromFileForDiff(codeView)
            const resolveBaseCommitID = resolveDiffRev(
                {
                    repoName: state.baseRawRepoName,
                    differentialID: state.differentialID,
                    diffID: (state.leftDiffID || state.diffID)!,
                    leftDiffID: state.leftDiffID,
                    useDiffForBase: Boolean(state.leftDiffID), // if ?vs and base is not `on` i.e. the initial commit)
                    useBaseForDiff: false,
                    filePath: baseFilePath || filePath,
                    isBase: true,
                },
                requestGraphQL
            ).pipe(
                map(({ commitID, stagingRepoName }) => ({
                    baseCommitID: commitID,
                    baseRawRepoName: stagingRepoName || state.baseRawRepoName,
                })),
                catchError(err => {
                    throw err
                })
            )
            const resolveHeadCommitID = resolveDiffRev(
                {
                    repoName: state.headRawRepoName,
                    differentialID: state.differentialID,
                    diffID: state.diffID!,
                    leftDiffID: state.leftDiffID,
                    useDiffForBase: false,
                    useBaseForDiff: false,
                    filePath,
                    isBase: false,
                },
                requestGraphQL
            ).pipe(
                map(({ commitID, stagingRepoName }) => ({
                    commitID,
                    rawRepoName: stagingRepoName || state.headRawRepoName,
                })),
                catchError(err => {
                    throw err
                })
            )
            return zip(resolveBaseCommitID, resolveHeadCommitID).pipe(
                map(
                    ([{ baseCommitID, baseRawRepoName }, { commitID, rawRepoName }]): FileInfo => ({
                        ...state,
                        baseCommitID,
                        commitID,
                        filePath,
                        baseFilePath,
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
    from(getPhabricatorState(window.location, requestGraphQL)).pipe(
        filter((state): state is DiffusionState => state !== null && state.mode === PhabricatorMode.Diffusion)
    )
