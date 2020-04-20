import { TextDocumentIdentifier } from '../../../../shared/src/api/client/types/textDocument'
import { TextDocumentPositionParams } from '../../../../shared/src/api/protocol'
import { AbsoluteRepoFilePosition, FileSpec, RepoSpec, ResolvedRevSpec } from '../../../../shared/src/util/url'

export const toTextDocumentIdentifier = (pos: RepoSpec & ResolvedRevSpec & FileSpec): TextDocumentIdentifier => ({
    uri: `git://${pos.repoName}?${pos.commitID}#${pos.filePath}`,
})

export const toTextDocumentPositionParams = (pos: AbsoluteRepoFilePosition): TextDocumentPositionParams => ({
    textDocument: toTextDocumentIdentifier(pos),
    position: {
        character: pos.position.character - 1,
        line: pos.position.line - 1,
    },
})
