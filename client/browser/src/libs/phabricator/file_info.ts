import { from, Observable, zip } from 'rxjs'
import { catchError, filter, map, switchMap } from 'rxjs/operators'
import { DifferentialState, DiffusionState, PhabricatorMode, RevisionState } from '.'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { FileInfo } from '../code_intelligence'
import { resolveDiffRev } from './backend'
import { getFilepathFromFileForDiff, getFilePathFromFileForRevision } from './scrape'
import { getPhabricatorState } from './util'

export const resolveRevisionFileInfo = (
    codeView: HTMLElement,
    queryGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    from(getPhabricatorState(window.location, queryGraphQL)).pipe(
        filter((state): state is RevisionState => state !== null && state.mode === PhabricatorMode.Revision),
        map(state => ({
            repoName: state.repoName,
            commitID: state.headCommitID,
            baseCommitID: state.baseCommitID,
        })),
        map(info => ({
            ...info,
            filePath: getFilePathFromFileForRevision(codeView),
        }))
    )

export const resolveDiffFileInfo = (
    codeView: HTMLElement,
    queryGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    from(getPhabricatorState(window.location, queryGraphQL)).pipe(
        filter(state => state !== null && state.mode === PhabricatorMode.Differential),
        map(state => state as DifferentialState),
        map(state => {
            const { filePath, baseFilePath } = getFilepathFromFileForDiff(codeView)

            return {
                ...state,
                filePath,
                baseFilePath,
            }
        }),
        switchMap(info => {
            const resolveBaseCommitID = resolveDiffRev(
                {
                    repoName: info.baseRepoName,
                    differentialID: info.differentialID,
                    diffID: (info.leftDiffID || info.diffID)!,
                    leftDiffID: info.leftDiffID,
                    useDiffForBase: Boolean(info.leftDiffID), // if ?vs and base is not `on` i.e. the initial commit)
                    useBaseForDiff: false,
                    filePath: info.baseFilePath || info.filePath,
                    isBase: true,
                },
                queryGraphQL
            ).pipe(
                map(({ commitID, stagingRepoName }) => ({
                    baseCommitID: commitID,
                    baseRepoName: stagingRepoName || info.baseRepoName,
                })),
                catchError(err => {
                    throw err
                })
            )

            const resolveHeadCommitID = resolveDiffRev(
                {
                    repoName: info.headRepoName,
                    differentialID: info.differentialID,
                    diffID: info.diffID!,
                    leftDiffID: info.leftDiffID,
                    useDiffForBase: false,
                    useBaseForDiff: false,
                    filePath: info.filePath,
                    isBase: false,
                },
                queryGraphQL
            ).pipe(
                map(({ commitID, stagingRepoName }) => ({
                    headCommitID: commitID,
                    headRepoName: stagingRepoName || info.headRepoName,
                })),
                catchError(err => {
                    throw err
                })
            )

            return zip(resolveBaseCommitID, resolveHeadCommitID).pipe(
                map(([{ baseCommitID, baseRepoName }, { headCommitID, headRepoName }]) => ({
                    baseCommitID,
                    headCommitID,
                    ...info,
                    baseRepoName,
                    headRepoName,
                }))
            )
        }),
        map(info => ({
            repoName: info.headRepoName,
            filePath: info.filePath,
            commitID: info.headCommitID,
            rev: info.headRev,
            baseRepoName: info.baseRepoName,
            baseFilePath: info.baseFilePath || info.filePath,
            baseCommitID: info.baseCommitID,
            baseRev: info.baseRev,
        }))
    )

export const resolveDiffusionFileInfo = (
    codeView: HTMLElement,
    queryGraphQL: PlatformContext['requestGraphQL']
): Observable<FileInfo> =>
    from(getPhabricatorState(window.location, queryGraphQL)).pipe(
        filter((state): state is DiffusionState => state !== null && state.mode === PhabricatorMode.Diffusion)
    )
