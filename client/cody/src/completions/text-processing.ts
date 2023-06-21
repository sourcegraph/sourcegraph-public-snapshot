import * as vscode from 'vscode'

export const OPENING_CODE_TAG = '<CODE5711>'
export const CLOSING_CODE_TAG = '</CODE5711>'

export function extractFromCodeBlock(completion: string): string {
    if (completion.includes(OPENING_CODE_TAG)) {
        // TODO(valery): use logger here instead.
        // console.error('invalid code completion response, should not contain opening tag <CODE5711>')
        return ''
    }

    const [result] = completion.split(CLOSING_CODE_TAG)

    return result.trimEnd()
}

const INDENTATION_REGEX = /^[\t ]*/
/**
 * Counts space or tabs in the beginning of a line.
 *
 * Since Cody can sometimes respond in a mix of tab and spaces, this function
 * normalizes the whitespace first using the currently enabled tabSize option.
 */
export function indentation(line: string): number {
    const tabSize = vscode.window.activeTextEditor
        ? // tabSize is always resolved to a number when accessing the property
          (vscode.window.activeTextEditor.options.tabSize as number)
        : 2

    const regex = line.match(INDENTATION_REGEX)
    if (regex) {
        const whitespace = regex[0]
        return [...whitespace].reduce((p, c) => p + (c === '\t' ? tabSize : 1), 0)
    }

    return 0
}

const BAD_COMPLETION_START = /^(\p{Emoji_Presentation}|\u{200B}|\+ |- |. )+(\s)+/u
export function fixBadCompletionStart(completion: string): string {
    if (BAD_COMPLETION_START.test(completion)) {
        return completion.replace(BAD_COMPLETION_START, '')
    }

    return completion
}

export interface TrimmedString {
    trimmed: string
    leadSpace: string
    rearSpace: string
}

export interface PrefixComponents {
    head: TrimmedString
    tail: TrimmedString
    overlap?: string
}

// Split string into head and tail. The tail is at most the last 2 non-empty lines of the snippet
export function getHeadAndTail(s: string): PrefixComponents {
    const lines = s.split('\n')
    const tailThreshold = 2

    let nonEmptyCount = 0
    let tailStart = -1
    for (let i = lines.length - 1; i >= 0; i--) {
        if (lines[i].trim().length > 0) {
            nonEmptyCount++
        }
        if (nonEmptyCount >= tailThreshold) {
            tailStart = i
            break
        }
    }

    if (tailStart === -1) {
        return { head: trimSpace(s), tail: trimSpace(s), overlap: s }
    }

    return { head: trimSpace(lines.slice(0, tailStart).join('\n')), tail: trimSpace(lines.slice(tailStart).join('\n')) }
}

function trimSpace(s: string): TrimmedString {
    const trimmed = s.trim()
    const headEnd = s.indexOf(trimmed)
    return { trimmed, leadSpace: s.slice(0, headEnd), rearSpace: s.slice(headEnd + trimmed.length) }
}

export function trimUntilSuffix(insertion: string, suffix: string): string {
    insertion = insertion.trimEnd()
    let firstNonEmptySuffixLine = ''
    for (const line of suffix.split('\n')) {
        if (line.trim().length > 0) {
            firstNonEmptySuffixLine = line
            break
        }
    }
    if (firstNonEmptySuffixLine.length === 0) {
        return insertion
    }

    const insertionLines = insertion.split('\n')
    let insertionEnd = insertionLines.length
    for (let i = 0; i < insertionLines.length; i++) {
        const line = insertionLines[i]
        if (line === firstNonEmptySuffixLine) {
            insertionEnd = i
            break
        }
    }
    return insertionLines.slice(0, insertionEnd).join('\n')
}
