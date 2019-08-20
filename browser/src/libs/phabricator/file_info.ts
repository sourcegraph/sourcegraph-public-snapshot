import { Observable, zip } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { FileInfo } from '../code_intelligence'
import { PhabricatorMode } from '.'
import { queryConduit as queryConduitAPI, resolveDiffRev } from './backend'
import { getFilepathFromFileForDiff, getFilePathFromFileForRevision } from './scrape'
import { getPhabricatorState } from './util'

export const resolveRevisionFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit = queryConduitAPI
): Observable<FileInfo> =>
    getPhabricatorState(window.location, requestGraphQL, queryConduit).pipe(
        map(
            (state): FileInfo => {
                if (state.mode !== PhabricatorMode.Revision) {
                    throw new Error(
                        `Unexpected Phabricator state for resolveRevisionFileInfo, PhabricatorMode: ${state.mode}`
                    )
                }
                const { rawRepoName, headCommitID, baseCommitID } = state
                return {
                    rawRepoName,
                    commitID: headCommitID,
                    baseCommitID,
                    filePath: getFilePathFromFileForRevision(codeView),
                }
            }
        )
    )

export const resolveDiffFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit = queryConduitAPI
): Observable<FileInfo> =>
    getPhabricatorState(window.location, requestGraphQL, queryConduit).pipe(
        switchMap(state => {
            if (state.mode !== PhabricatorMode.Differential) {
                throw new Error(`Unexpected PhabricatorState for resolveDiffFileInfo, PhabricatorMode: ${state.mode}`)
            }
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
                requestGraphQL,

                queryConduit
            ).pipe(
                map(
                    ({
                        commitID,
                        stagingRepoName,
                        isStagingCommit,
                    }): Pick<FileInfo, 'baseCommitID' | 'baseRev' | 'baseRawRepoName'> => ({
                        baseCommitID: commitID,
                        // Only keep the rev if this is not a staging commit,
                        // or if the staging repo is synced to the Sourcegraph instance.
                        baseRev: !isStagingCommit || stagingRepoName !== undefined ? state.baseRev : undefined,
                        baseRawRepoName: stagingRepoName || state.baseRawRepoName,
                    })
                )
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
                requestGraphQL,

                queryConduit
            ).pipe(
                map(
                    ({
                        commitID,
                        stagingRepoName,
                        isStagingCommit,
                    }): Pick<FileInfo, 'commitID' | 'rev' | 'rawRepoName'> => ({
                        commitID,
                        // Only keep the rev if this is not a staging commit,
                        // or if the staging repo is synced to the Sourcegraph instance.
                        rev: !isStagingCommit || stagingRepoName !== undefined ? state.headRev : undefined,
                        rawRepoName: stagingRepoName || state.headRawRepoName,
                    })
                )
            )
            return zip(resolveBaseCommitID, resolveHeadCommitID).pipe(
                map(
                    ([baseInfo, headInfo]): FileInfo => ({
                        ...baseInfo,
                        ...headInfo,
                        baseFilePath,
                        filePath,
                    })
                )
            )
        })
    )

export const resolveDiffusionFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit = queryConduitAPI
): Observable<FileInfo> =>
    getPhabricatorState(window.location, requestGraphQL, queryConduit).pipe(
        map(
            (state): FileInfo => {
                if (state.mode !== PhabricatorMode.Diffusion) {
                    throw new Error(
                        `Unexpected PhabricatorState for resolveDiffusionFileInfo, PhabricatorMode: ${state.mode}`
                    )
                }
                const { filePath, commitID, rawRepoName } = state
                return {
                    filePath,
                    commitID,
                    rawRepoName,
                }
            }
        )
    )
