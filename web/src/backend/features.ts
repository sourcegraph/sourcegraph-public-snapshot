import { Observable } from 'rxjs'
import { HoverMerged } from '../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { FileSpec, UIPositionSpec, RepoSpec, ResolvedRevisionSpec } from '../../../shared/src/util/url'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'

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
    return extensionsController.services.textDocumentHover.getHover({
        textDocument: { uri: `git://${context.repoName}?${context.commitID}#${context.filePath}` },
        position: {
            character: context.position.character - 1,
            line: context.position.line - 1,
        },
    })
}
