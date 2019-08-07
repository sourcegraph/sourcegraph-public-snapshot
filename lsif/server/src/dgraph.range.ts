import * as lsp from 'vscode-languageserver'

/**
 * A flattened LSP range.
 */
export interface FlatRange {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

/**
 * Convert an LSP range into a FlatRange.
 */
export function flattenRange(range: lsp.Range): FlatRange {
    return {
        startLine: range.start.line,
        startCharacter: range.start.character,
        endLine: range.end.line,
        endCharacter: range.end.character,
    }
}

/**
 * Convert a FlatRange into an LSP range.
 */
export function unflattenRange(flatRange: FlatRange): lsp.Range {
    return {
        start: { line: flatRange.startLine, character: flatRange.startCharacter },
        end: { line: flatRange.endLine, character: flatRange.endCharacter },
    }
}
