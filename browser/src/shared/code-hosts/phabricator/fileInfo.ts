import { Observable, zip } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { FileInfo, DiffOrBlobInfo } from '../shared/codeHost'
import { PhabricatorMode } from '.'
import { queryConduitHelper, resolveDiffRev } from './backend'
import { getFilepathFromFileForDiff, getFilePathFromFileForRevision } from './scrape'
import { getPhabricatorState } from './util'

export const resolveRevisionFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit = queryConduitHelper
): Observable<DiffOrBlobInfo> =>
    getPhabricatorState(window.location, requestGraphQL, queryConduit).pipe(
        map(
            (state): DiffOrBlobInfo => {
                if (state.mode !== PhabricatorMode.Revision) {
                    throw new Error(
                        `Unexpected Phabricator state for resolveRevisionFileInfo, PhabricatorMode: ${state.mode}`
                    )
                }
                const { rawRepoName, headCommitID, baseCommitID } = state
                const filePath = getFilePathFromFileForRevision(codeView)
                return {
                    base: {
                        rawRepoName,
                        filePath,
                        commitID: baseCommitID,
                    },
                    head: {
                        rawRepoName,
                        filePath,
                        commitID: headCommitID,
                    },
                }
            }
        )
    )

export const resolveDiffFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit = queryConduitHelper
): Observable<DiffOrBlobInfo> =>
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
                    ({ commitID, stagingRepoName }): Pick<FileInfo, 'commitID' | 'rawRepoName'> => ({
                        commitID,
                        rawRepoName: stagingRepoName || state.baseRawRepoName,
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
                    ({ commitID, stagingRepoName }): Pick<FileInfo, 'commitID' | 'rawRepoName'> => ({
                        commitID,
                        rawRepoName: stagingRepoName || state.headRawRepoName,
                    })
                )
            )
            return zip(resolveBaseCommitID, resolveHeadCommitID).pipe(
                map(
                    ([baseInfo, headInfo]): DiffOrBlobInfo => ({
                        base: {
                            rawRepoName: baseInfo.rawRepoName,
                            filePath: baseFilePath || filePath,
                            commitID: baseInfo.commitID,
                        },
                        head: {
                            rawRepoName: headInfo.rawRepoName,
                            filePath,
                            commitID: headInfo.commitID,
                        },
                    })
                )
            )
        })
    )

export const resolveDiffusionFileInfo = (
    codeView: HTMLElement,
    requestGraphQL: PlatformContext['requestGraphQL'],
    queryConduit = queryConduitHelper
): Observable<DiffOrBlobInfo> =>
    getPhabricatorState(window.location, requestGraphQL, queryConduit).pipe(
        map(
            (state): DiffOrBlobInfo => {
                if (state.mode !== PhabricatorMode.Diffusion) {
                    throw new Error(
                        `Unexpected PhabricatorState for resolveDiffusionFileInfo, PhabricatorMode: ${state.mode}`
                    )
                }
                const { filePath, commitID, rawRepoName } = state
                return {
                    blob: {
                        filePath,
                        commitID,
                        rawRepoName,
                    },
                }
            }
        )
    )
