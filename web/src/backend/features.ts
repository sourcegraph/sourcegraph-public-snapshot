import { Observable } from 'rxjs'
import { HoverMerged } from '../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { FileSpec, PositionSpec, RepoSpec, ResolvedRevSpec } from '../../../shared/src/util/url'

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(
    ctx: RepoSpec & ResolvedRevSpec & FileSpec & PositionSpec,
    { extensionsController }: ExtensionsControllerProps
): Observable<HoverMerged | null> {
    return extensionsController.services.textDocumentHover.getHover({
        textDocument: { uri: `git://${ctx.repoName}?${ctx.commitID}#${ctx.filePath}` },
        position: {
            character: ctx.position.character - 1,
            line: ctx.position.line - 1,
        },
    })
}
