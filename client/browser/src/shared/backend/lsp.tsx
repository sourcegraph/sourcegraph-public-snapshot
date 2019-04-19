import { Location } from '@sourcegraph/extension-api-types'
import { from, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { HoverMerged } from '../../../../../shared/src/api/client/types/hover'
import { TextDocumentIdentifier } from '../../../../../shared/src/api/client/types/textDocument'
import { TextDocumentPositionParams } from '../../../../../shared/src/api/protocol'
import { Controller } from '../../../../../shared/src/extensions/controller'
import { AbsoluteRepoFilePosition, FileSpec, RepoSpec, ResolvedRevSpec } from '../../../../../shared/src/util/url'

interface SimpleProviderFns {
    getHover: (pos: AbsoluteRepoFilePosition) => Observable<HoverMerged | null>
    fetchDefinition: (pos: AbsoluteRepoFilePosition) => Observable<Location | Location[] | null>
}

export const toTextDocumentIdentifier = (pos: RepoSpec & ResolvedRevSpec & FileSpec): TextDocumentIdentifier => ({
    uri: `git://${pos.repoName}?${pos.commitID}#${pos.filePath}`,
})

const toTextDocumentPositionParams = (pos: AbsoluteRepoFilePosition): TextDocumentPositionParams => ({
    textDocument: toTextDocumentIdentifier(pos),
    position: {
        character: pos.position.character - 1,
        line: pos.position.line - 1,
    },
})

export const createLSPFromExtensions = (extensionsController: Controller): SimpleProviderFns => ({
    // Use from() to suppress rxjs type incompatibilities between different minor versions of rxjs in
    // node_modules/.
    getHover: pos =>
        from(extensionsController.services.textDocumentHover.getHover(toTextDocumentPositionParams(pos))).pipe(
            map(hover => (hover === null ? HoverMerged.from([]) : hover))
        ),
    fetchDefinition: pos =>
        from(extensionsController.services.textDocumentDefinition.getLocations(toTextDocumentPositionParams(pos))).pipe(
            switchMap(locations => locations)
        ) as Observable<Location | Location[] | null>,
})
