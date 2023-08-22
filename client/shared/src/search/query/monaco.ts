// A collection of Monaco related helper functions
import type * as Monaco from 'monaco-editor'

import type { CharacterRange } from './token'

/**
 * Converts a zero-indexed, possibly multiline, offset {@link CharacterRange} to a {@link Monaco.IRange} line-and-column range.
 */
export const toMonacoRange = ({ start, end }: CharacterRange, textModel: Monaco.editor.ITextModel): Monaco.IRange => {
    const startPosition = textModel.getPositionAt(start)
    const endPosition = textModel.getPositionAt(end)
    return {
        startLineNumber: startPosition.lineNumber,
        endLineNumber: endPosition.lineNumber,
        startColumn: startPosition.column,
        endColumn: endPosition.column,
    }
}

/**
 * Converts a zero-indexed, single-line {@link CharacterRange} to a Monaco {@link IRange}.
 */
export const toMonacoSingleLineRange = ({ start, end }: CharacterRange): Monaco.IRange => ({
    startLineNumber: 1,
    endLineNumber: 1,
    startColumn: start + 1,
    endColumn: end + 1,
})
