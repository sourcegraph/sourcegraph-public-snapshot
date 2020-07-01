import { Observable, from, concat } from 'rxjs'
import { HoverMerged } from '../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { FileSpec, UIPositionSpec, RepoSpec, ResolvedRevisionSpec } from '../../../shared/src/util/url'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { switchMap } from 'rxjs/operators'
import { wrapRemoteObservable } from '../../../shared/src/api/client/api/common'
import { DocumentHighlight } from 'sourcegraph'

/**
 * Fetches hover information for the given location.
 *
 * @param context the location
 * @returns hover for the location
 */
export function getHover(
    context: RepoSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec,
    { extensionsController }: ExtensionsControllerProps
): Observable<MaybeLoadingResult<HoverMerged | null>> {
    return concat(
        [{ isLoading: true, result: null }],
        from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHost =>
                wrapRemoteObservable(
                    extensionHost.getHover({
                        textDocument: {
                            uri: `git://${context.repoName}?${context.commitID}#${context.filePath}`,
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
 * TODO - document
 */
export function getDocumentHighlights(
    context: RepoSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec,
    { extensionsController }: ExtensionsControllerProps
): Observable<MaybeLoadingResult<DocumentHighlight[] | null>> {
    return concat(
        [{ isLoading: true, result: null }],
        from(extensionsController.extHostAPI).pipe(
            switchMap(extensionHost =>
                wrapRemoteObservable(
                    extensionHost.getDocumentHighlights({
                        textDocument: {
                            uri: `git://${context.repoName}?${context.commitID}#${context.filePath}`,
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
