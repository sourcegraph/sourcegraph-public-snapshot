import detectIndent from 'detect-indent'

import { PrefixComponents, getEditorTabSize, indentation } from './text-processing'

const BRACKET_PAIR = {
    '(': ')',
    '[': ']',
    '{': '}',
} as const
const OPENING_BRACKET_REGEX = /([([{])$/
export function detectMultilineMode(
    prefix: string,
    prevNonEmptyLine: string,
    sameLinePrefix: string,
    sameLineSuffix: string,
    languageId: string,
    enableExtendedTriggers: boolean
): null | 'block' {
    const config = getLanguageConfig(languageId)
    if (!config) {
        return null
    }

    if (enableExtendedTriggers && sameLinePrefix.match(OPENING_BRACKET_REGEX)) {
        return 'block'
    }

    if (
        sameLinePrefix.trim() === '' &&
        sameLineSuffix.trim() === '' &&
        // Only trigger multiline suggestions for the beginning of blocks
        prefix.trim().at(prefix.trim().length - config.blockStart.length) === config.blockStart &&
        // Only trigger multiline suggestions when the new current line is indented
        indentation(prevNonEmptyLine) < indentation(sameLinePrefix)
    ) {
        return 'block'
    }

    return null
}

// Detect if completion starts with a space followed by any non-space character.
export const ODD_INDENTATION_REGEX = /^ [^ ]/
export function checkOddIndentation(completion: string, prefix: PrefixComponents): boolean {
    return completion.length > 0 && ODD_INDENTATION_REGEX.test(completion) && prefix.tail.rearSpace.length > 0
}

/**
 * Adjusts the indentation of a multiline completion to match the current editor indentation.
 */
function adjustIndentation(text: string, originalIndent: number, newIndent: number): string {
    const lines = text.split('\n')

    return lines
        .map(line => {
            let spaceCount = 0
            for (const char of line) {
                if (char === ' ') {
                    spaceCount++
                } else {
                    break
                }
            }

            const indentLevel = spaceCount / originalIndent

            if (Number.isInteger(indentLevel)) {
                const newIndentStr = ' '.repeat(indentLevel * newIndent)
                return line.replace(/^ +/, newIndentStr)
            }

            // The line has a non-standard number of spaces at the start, leave it unchanged
            return line
        })
        .join('\n')
}

function ensureSameOrLargerIndentation(completion: string): string {
    const indentAmount = detectIndent(completion).amount
    const editorTabSize = getEditorTabSize()

    if (editorTabSize > indentAmount) {
        return adjustIndentation(completion, indentAmount, editorTabSize)
    }

    return completion
}

/**
 * If a completion starts with an opening bracket and a suffix starts with
 * the corresponding closing bracket, we include the last closing bracket of the completion.
 * E.g., function foo(__CURSOR__)
 *
 * We can do this because we know that the existing block is already closed, which means that
 * new blocks need to be closed separately.
 * E.g. function foo() { console.log('hello') }
 */
function shouldIncludeClosingLineBasedOnBrackets(
    prefixIndentationWithFirstCompletionLine: string,
    suffix: string
): boolean {
    const matches = prefixIndentationWithFirstCompletionLine.match(OPENING_BRACKET_REGEX)

    if (matches && matches.length > 0) {
        const openingBracket = matches[0] as keyof typeof BRACKET_PAIR
        const closingBracket = BRACKET_PAIR[openingBracket]

        return Boolean(openingBracket) && suffix.startsWith(closingBracket)
    }

    return false
}

/**
 * Only include a closing line (e.g. `}`) if the block is empty yet if the block is already closed.
 * We detect this by looking at the indentation of the next non-empty line.
 */
function shouldIncludeClosingLine(prefixIndentationWithFirstCompletionLine: string, suffix: string): boolean {
    const includeClosingLineBasedOnBrackets = shouldIncludeClosingLineBasedOnBrackets(
        prefixIndentationWithFirstCompletionLine,
        suffix
    )
    const startIndent = indentation(prefixIndentationWithFirstCompletionLine)

    const firstNewLineIndex = suffix.indexOf('\n') + 1
    const nextNonEmptyLine =
        suffix
            .slice(firstNewLineIndex)
            .split('\n')
            .find(line => line.trim().length > 0) ?? ''

    return indentation(nextNonEmptyLine) < startIndent || includeClosingLineBasedOnBrackets
}

export function truncateMultilineCompletion(
    completion: string,
    prefix: string,
    suffix: string,
    languageId: string
): string {
    const config = getLanguageConfig(languageId)
    if (!config) {
        return completion
    }

    // Ensure that the completion has the same or larger indentation
    // because we rely on the indentation size to cut off the completion.
    // TODO: add unit tests for this case. We need to update the indentation logic
    // used in unit tests for code samples.
    const indentedCompletion = ensureSameOrLargerIndentation(completion)
    const lines = indentedCompletion.split('\n')

    // We use a whitespace counting approach to finding the end of the
    // completion. To find an end, we look for the first line that is below the
    // start scope of the completion ( calculated by the number of leading
    // spaces or tabs)
    const prefixLastNewline = prefix.lastIndexOf('\n')
    const prefixIndentationWithFirstCompletionLine = prefix.slice(prefixLastNewline + 1)
    const startIndent = indentation(prefixIndentationWithFirstCompletionLine)
    const hasEmptyCompletionLine = prefixIndentationWithFirstCompletionLine.trim() === ''

    // Normalize responses that start with a newline followed by the exact
    // indentation of the first line.
    if (lines.length > 1 && lines[0] === '' && indentation(lines[1]) === startIndent) {
        lines.shift()
        lines[0] = lines[0].trimStart()
    }

    const includeClosingLine = shouldIncludeClosingLine(prefixIndentationWithFirstCompletionLine, suffix)

    let cutOffIndex = lines.length
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i]

        if (i === 0 || line === '' || config.blockElseTest.test(line)) {
            continue
        }

        if (
            (indentation(line) <= startIndent && !hasEmptyCompletionLine) ||
            (indentation(line) < startIndent && hasEmptyCompletionLine)
        ) {
            // When we find the first block below the start indentation, only
            // include it if it is an end block
            if (includeClosingLine && config.blockEnd && line.trim().startsWith(config.blockEnd)) {
                cutOffIndex = i + 1
            } else {
                cutOffIndex = i
            }
            break
        }
    }

    return lines.slice(0, cutOffIndex).join('\n')
}

interface LanguageConfig {
    blockStart: string
    blockElseTest: RegExp
    blockEnd: string | null
}
function getLanguageConfig(languageId: string): LanguageConfig | null {
    switch (languageId) {
        case 'c':
        case 'cpp':
        case 'csharp':
        case 'go':
        case 'java':
        case 'javascript':
        case 'javascriptreact':
        case 'typescript':
        case 'typescriptreact':
            return {
                blockStart: '{',
                blockElseTest: /^[\t ]*} else/,
                blockEnd: '}',
            }
        case 'python': {
            return {
                blockStart: ':',
                blockElseTest: /^[\t ]*(elif |else:)/,
                blockEnd: null,
            }
        }
        default:
            return null
    }
}
