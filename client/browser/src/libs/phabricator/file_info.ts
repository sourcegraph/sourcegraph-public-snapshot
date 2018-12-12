import { from, Observable, zip } from 'rxjs'
import { catchError, filter, map, switchMap } from 'rxjs/operators'
import { DifferentialState, DiffusionState, PhabricatorMode } from '.'
import { fetchBlobContentLines } from '../../shared/repo/backend'
import { FileInfo } from '../code_intelligence'
import { ensureRevisionsAreCloned } from '../code_intelligence/util/file_info'
import { resolveDiffRev } from './backend'
import { getFilepathFromFile, getPhabricatorState } from './util'

export const resolveDiffFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    from(getPhabricatorState(window.location)).pipe(
        filter(state => state !== null && state.mode === PhabricatorMode.Differential),
        map(state => state as DifferentialState),
        map(state => {
            const { filePath, baseFilePath } = getFilepathFromFile(codeView)

            return {
                ...state,
                filePath,
                baseFilePath,
            }
        }),
        switchMap(info => {
            const resolveBaseCommitID = resolveDiffRev({
                repoName: info.baseRepoName,
                differentialID: info.differentialID,
                diffID: (info.leftDiffID || info.diffID)!,
                leftDiffID: info.leftDiffID,
                useDiffForBase: Boolean(info.leftDiffID), // if ?vs and base is not `on` i.e. the initial commit)
                useBaseForDiff: false,
                filePath: info.baseFilePath || info.filePath,
                isBase: true,
            }).pipe(
                map(({ commitID, stagingRepoName }) => ({
                    baseCommitID: commitID,
                    baseRepoName: stagingRepoName || info.baseRepoName,
                })),
                catchError(err => {
                    throw err
                })
            )

            const resolveHeadCommitID = resolveDiffRev({
                repoName: info.headRepoName,
                differentialID: info.differentialID,
                diffID: info.diffID!,
                leftDiffID: info.leftDiffID,
                useDiffForBase: false,
                useBaseForDiff: false,
                filePath: info.filePath,
                isBase: false,
            }).pipe(
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
        switchMap(info => {
            const fetchingBaseFile = fetchBlobContentLines({
                repoName: info.baseRepoName,
                filePath: info.baseFilePath || info.filePath,
                commitID: info.baseCommitID,
                rev: info.baseRev,
            })

            const fetchingHeadFile = fetchBlobContentLines({
                repoName: info.headRepoName,
                filePath: info.filePath,
                commitID: info.headCommitID,
                rev: info.headRev,
            })

            return zip(fetchingBaseFile, fetchingHeadFile).pipe(
                map(([baseFileContent, headFileContent]) => ({ ...info, baseFileContent, headFileContent }))
            )
        }),
        map(info => ({
            repoName: info.headRepoName,
            filePath: info.filePath,
            commitID: info.headCommitID,
            rev: info.headRev,

            baseRepoName: info.baseRepoName,
            baseFilePath: info.baseFileContent ? info.baseFilePath || info.filePath : undefined,
            baseCommitID: info.baseCommitID,
            baseRev: info.baseRev,

            headHasFileContents: info.headFileContent.length > 0,
            baseHasFileContents: info.baseFileContent.length > 0,
        })),
        ensureRevisionsAreCloned
    )

export const resolveDiffusionFileInfo = (codeView: HTMLElement): Observable<FileInfo> =>
    from(getPhabricatorState(window.location)).pipe(
        filter(state => state !== null && state.mode === PhabricatorMode.Diffusion),
        map(state => state as DiffusionState)
    )
