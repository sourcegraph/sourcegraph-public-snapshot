import path from 'path'

import { JointRange, Position, TextDocument } from '../editor'
import { DocumentOffsets } from '../editor/offsets'

export interface Completion {
    prefix: string
    content: string
    stopReason?: string
}

export interface AutocompleteContext extends TextDocument {
    markdownLanguage: string

    prefix: JointRange
    suffix: JointRange

    prevLine: JointRange | null
    prevNonEmptyLine: JointRange | null
    nextNonEmptyLine: JointRange | null
}

/**
 * Get the current document context based on the cursor position in the current document.
 *
 * This function is meant to provide a context around the current position in the document,
 * including a prefix, a suffix, the previous line, the previous non-empty line, and the next non-empty line.
 * The prefix and suffix are obtained by looking around the current position up to a max length
 * defined by `maxPrefixLength` and `maxSuffixLength` respectively. If the length of the entire
 * document content in either direction is smaller than these parameters, the entire content will be used.
 */
export function getAutocompleteContext(
    document: TextDocument,
    position: Position,
    maxPrefixLength: number,
    maxSuffixLength: number
): AutocompleteContext {
    const offset = new DocumentOffsets(document.content)
    const posOffset = offset.offset(position)

    let prevNonEmptyLine = null
    for (let line = position.line - 1; line >= 0; line--) {
        if (offset.getLine(line).trim().length !== 0) {
            prevNonEmptyLine = offset.toJointRange(offset.getLineRange(line))
        }
    }

    let nextNonEmptyLine = null
    for (let line = position.line + 1; line < offset.lines.length; line++) {
        if (offset.getLine(line).trim().length !== 0) {
            nextNonEmptyLine = offset.toJointRange(offset.getLineRange(line))
        }
    }

    return {
        ...document,

        markdownLanguage: path.extname(document.uri),

        prefix: offset.toJointRange({
            start: offset.position(Math.max(0, posOffset - maxPrefixLength)),
            end: position,
        }),
        suffix: offset.toJointRange({
            start: position,
            end: offset.position(Math.min(posOffset + maxSuffixLength, document.content.length)),
        }),

        prevLine: position.line > 0 ? offset.toJointRange(offset.getLineRange(position.line - 1)) : null,

        prevNonEmptyLine,
        nextNonEmptyLine,
    }
}
