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

export function getAutocompleteContext(
    document: TextDocument,
    position: Position,
    maxPrefixLength: number,
    maxSuffixLength: number
): AutocompleteContext {
    const offset = new DocumentOffsets(document.content)
    const posOffset = offset.offset(position)

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
