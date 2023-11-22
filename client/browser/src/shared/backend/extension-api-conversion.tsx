import type { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { type AbsoluteRepoFilePosition, toURIWithPath } from '@sourcegraph/shared/src/util/url'

export const toTextDocumentPositionParameters = (
    position: AbsoluteRepoFilePosition
): TextDocumentPositionParameters => ({
    textDocument: {
        uri: toURIWithPath(position),
    },
    position: {
        character: position.position.character - 1,
        line: position.position.line - 1,
    },
})
