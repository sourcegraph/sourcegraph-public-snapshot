import { flatten } from 'lodash'

import { BlockCommentStyle } from '../language-specs/language-spec'

/**
 * The default regex for characters allowed in an identifier. It works well for
 * C-like languages (C/C++, C#, Java, etc.) but not for languages that allow
 * punctuation characters (e.g. Ruby).
 */
const DEFAULT_IDENT_CHAR_PATTERN = /\w/

/**
 * Extract the token that occurs on the given line at the given position. This will
 * scan the line around the current hover position trying to return the maximal set
 * of characters that appear like a symbol given the identifier pattern.
 *
 * @param args Parameter bag.
 */
export function findSearchToken({
    text,
    position,
    lineRegexes,
    blockCommentStyles,
    identCharPattern = DEFAULT_IDENT_CHAR_PATTERN,
}: {
    /** The text of the current document. */
    text: string
    /** The current hover position. */
    position: { line: number; character: number }
    /** The patterns that identify line comments. */
    lineRegexes: RegExp[]
    /** The patterns that identify block comments. */
    blockCommentStyles: BlockCommentStyle[]
    /** The pattern that identifies identifiers in this language. */
    identCharPattern?: RegExp
}): { searchToken: string; isString: boolean; isComment: boolean } | undefined {
    const lines = text.split('\n')
    const line = lines[position.line]
    if (line === undefined) {
        // Weird case where the position is bogus relative to the text
        return undefined
    }

    // Scan from the current hover position to the right while the characters
    // still match the identifier pattern. If no characters match the pattern
    // then we default to the end of the line.

    let end = line.length
    for (let index = position.character; index < line.length; index++) {
        if (!identCharPattern.test(line[index])) {
            end = index
            break
        }
    }

    // Scan from the current hover position to the left while the characters
    // still match the identifier pattern. If no characters match the pattern
    // then we default to the start of the line.

    let start = 0
    for (let index = position.character; index >= 0; index--) {
        if (!identCharPattern.test(line[index])) {
            start = index + 1
            break
        }
    }

    if (start >= end) {
        return undefined
    }

    return {
        searchToken: line.slice(start, end),
        isString: isInsideString({ line: lines[position.line], start, end }),
        isComment: isInsideComment({ lines, position, start, end, lineRegexes, blockCommentStyles }),
    }
}

/**
 * Determine if the identifier matched on the given line occurs within a string.
 *
 * @param args Parameter bag.
 */
function isInsideString({
    line,
    start,
    end,
}: {
    /** The line containing the identifier */
    line: string
    /** The offset of the identifier in the target line. */
    start: number
    /** The offset and length of the identifier in the target line. */
    end: number
}): boolean {
    return checkMatchIntersection(Array.from(line.matchAll(/'.*?'|".*?"/gs)), { start, end })
}

/**
 * Determine if the identifier matched on the given line occurs within a comment
 * defined by the given line comment and block comment regular expressions.
 *
 * @param args Parameter bag.
 */
function isInsideComment({
    lines,
    position,
    start,
    end,
    blockCommentStyles,
    lineRegexes,
}: {
    /** The text of the current document split into lines. */
    lines: string[]
    /** The current hover position. */
    position: { line: number }
    /** The offset of the identifier in the target line. */
    start: number
    /** The offset and length of the identifier in the target line. */
    end: number
    /** The patterns that identify line comments. */
    lineRegexes: RegExp[]
    /** The patterns that identify block comments. */
    blockCommentStyles: BlockCommentStyle[]
}): boolean {
    const line = lines[position.line]

    if (
        isInsideLineComment({ line, start, lineRegexes }) ||
        isInsideBlockComment({ lines, position, start, end, blockCommentStyles })
    ) {
        const searchToken = lines[position.line].slice(start, end)

        const blessedPatterns = [
            // looks like a function call
            new RegExp(`${searchToken}\\(`),
            // looks like a field projection
            new RegExp(`\\.${searchToken}`),
        ]

        return !blessedPatterns.some(pattern => pattern.test(line))
    }

    return false
}

/**
 * Determine if the identifier matched on the given line occurs within a comment
 * defined by the given line comment regular expressions.
 *
 * @param args Parameter bag.
 */
function isInsideLineComment({
    line,
    start,
    lineRegexes,
}: {
    /** The line containing the identifier */
    line: string
    /** The index where the identifier occurs on the line. */
    start: number
    /** The patterns that identify line comments. */
    lineRegexes: RegExp[]
}): boolean {
    // Determine if the token occurs after a comment on the same line
    return lineRegexes.some(lineRegex => {
        const match = line.match(lineRegex)
        if (!match) {
            return false
        }

        return match?.index !== undefined && match.index < start
    })
}

/**
 * How many lines of context to capture on each side of a identifier when checking
 * whether or not the user is within a comment. A value of 50 will search over 101
 * lines in total.
 */
const LINES_OF_CONTEXT = 50

/**
 * Determine if the identifier matched on the given line occurs within a comment
 * defined by the given line block comment style.
 *
 * @param args Parameter bag.
 */
function isInsideBlockComment({
    lines,
    position,
    start,
    end,
    blockCommentStyles,
}: {
    /** The text of the current document split into lines. */
    lines: string[]
    /** The current hover position. */
    position: { line: number }
    /** The offset of the identifier in the target line. */
    start: number
    /** The offset and length of the identifier in the target line. */
    end: number
    /** The patterns that identify block comments. */
    blockCommentStyles: BlockCommentStyle[]
}): boolean {
    const line = lines[position.line]
    const linesBefore = lines.slice(Math.max(position.line - LINES_OF_CONTEXT, 0), position.line)
    const linesAfter = lines.slice(position.line + 1, position.line + LINES_OF_CONTEXT + 1)

    // Search over multiple lines of text covering our identifier
    const context = flatten([linesBefore, [line], linesAfter]).join('\n')

    // Determine how many characters in the context we skip before landing on our line
    const offset = linesBefore.reduce((accumulator, line) => accumulator + line.length, 0) + linesBefore.length

    // Match all commented blocks in the given block of text. We know
    // the range of the target identifier in this text: if it's covered
    // in a match's range then it is nested inside of a comment.
    return blockCommentStyles.some(block =>
        checkMatchIntersection(
            Array.from(context.matchAll(new RegExp(`${block.startRegex.source}.*?${block.endRegex.source}`, 'gs'))),
            { start: start + offset, end: end + offset }
        )
    )
}

/**
 * Determine if any of the matches in the given array cover the given range.
 */
function checkMatchIntersection(matches: RegExpMatchArray[], range: { start: number; end: number }): boolean {
    return matches.some(
        match => match.index !== undefined && match.index <= range.start && match.index + match[0].length >= range.end
    )
}
