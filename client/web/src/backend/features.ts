import { Remote } from 'comlink'
import { Observable, from, concat } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { DocumentHighlight } from 'sourcegraph'

import { HoverMerged } from '@sourcegraph/client-api'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { memoizeObservable } from '@sourcegraph/common'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import {
    FileSpec,
    UIPositionSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    toURIWithPath,
    toRootURI,
} from '@sourcegraph/shared/src/util/url'

/**
 * Fetches hover information for the given location.
 *
 * @param context the location
 * @returns hover for the location
 */
export function getHover(
    context: RepoSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec,
    { extensionsController }: ExtensionsControllerProps<'extHostAPI'>
): Observable<MaybeLoadingResult<HoverMerged | null>> {
    return concat(
        [{ isLoading: true, result: null }],
        from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHost =>
                wrapRemoteObservable(
                    extensionHost.getHover({
                        textDocument: {
                            uri: toURIWithPath(context),
                        },
                        position: {
                            character: context.position.character - 1,
                            line: context.position.line - 1,
                        },
                    })
                )
            )
        )
    )
}

/**
 * Fetches document highlight information for the given location.
 *
 * @param context the location
 * @returns document highlights for the location
 */
export function getDocumentHighlights(
    context: RepoSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec,
    { extensionsController }: ExtensionsControllerProps<'extHostAPI'>
): Observable<DocumentHighlight[]> {
    return concat(
        [[]],
        from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHost =>
                wrapRemoteObservable(
                    extensionHost.getDocumentHighlights({
                        textDocument: {
                            uri: toURIWithPath(context),
                        },
                        position: {
                            character: context.position.character - 1,
                            line: context.position.line - 1,
                        },
                    })
                )
            )
        )
    )
}

/**
 * Fetches file decorations
 */
export const getFileDecorations = ({
    extensionsController,
    ...parameters
}: {
    files: { url: string; isDirectory: boolean; name: string; path: string }[]
    /** uri of node from which this request is made. Used to construct cache key  */
    parentNodeUri: string
} & ExtensionsControllerProps<'extHostAPI'> &
    RepoSpec &
    ResolvedRevisionSpec): Observable<FileDecorationsByPath> =>
    from(extensionsController.extHostAPI).pipe(
        switchMap(extensionHost => getFileDecorationsFromHost({ ...parameters, extensionHost }))
    )

const getFileDecorationsFromHost = memoizeObservable(
    ({
        files,
        extensionHost,
        commitID,
        repoName,
    }: {
        files: { url: string; isDirectory: boolean; name: string; path: string }[]
        /** uri of node from which this request is made. Used to construct cache key  */
        parentNodeUri: string
        extensionHost: Remote<FlatExtensionHostAPI>
    } & RepoSpec &
        ResolvedRevisionSpec) =>
        wrapRemoteObservable(
            extensionHost.getFileDecorations({
                uri: toRootURI({ repoName, commitID }),
                files: files.map(file => ({
                    ...file,
                    uri: toURIWithPath({ repoName, filePath: file.path, commitID }),
                })),
            })
        ),
    ({ parentNodeUri, files }) => `parentNodeUri:${parentNodeUri} files:${files.length}`
)
