import { TextDocumentIdentifier } from '@sourcegraph/shared/src/api/client/types/textDocument'
import { TextDocumentPositionParameters } from '@sourcegraph/shared/src/api/protocol'
import {
    AbsoluteRepoFilePosition,
    FileSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    toURIWithPath,
} from '@sourcegraph/shared/src/util/url'

export const toTextDocumentIdentifier = (
    position: RepoSpec & ResolvedRevisionSpec & FileSpec
): TextDocumentIdentifier => ({
    uri: toURIWithPath(position),
})

export const toTextDocumentPositionParameters = (
    position: AbsoluteRepoFilePosition
): TextDocumentPositionParameters => ({
    textDocument: toTextDocumentIdentifier(position),
    position: {
        character: position.position.character - 1,
        line: position.position.line - 1,
    },
})
