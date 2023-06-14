import { truncateMultilineCompletion } from './multiline'

const BAD_COMPLETION_START = /^(\p{Emoji_Presentation}|\u{200B}|\+ |- |. )+(\s)+/u

export function postProcess({
    prefix,
    suffix,
    languageId,
    multiline,
    completion,
}: {
    prefix: string
    suffix: string
    languageId: string
    multiline: boolean
    completion: string
}): null | string {
    // Extract a few common parts for the processing
    const currentLinePrefix = prefix.slice(prefix.lastIndexOf('\n') + 1)
    const firstNlInSuffix = suffix.indexOf('\n') + 1
    const nextNonEmptyLine =
        suffix
            .slice(firstNlInSuffix)
            .split('\n')
            .find(line => line.trim().length > 0) ?? ''

    // Single-line completions should be trimmed to only return one line
    if (!multiline) {
        completion = completion.slice(0, completion.indexOf('\n'))
    }

    // Sometimes Claude emits a single space in the completion. We call this an "odd indentation"
    // completion and try to fix the response.
    let hasOddIndentation = false
    if (
        completion.length > 0 &&
        /^ [^ ]/.test(completion) &&
        prefix.length > 0 &&
        (prefix.endsWith(' ') || prefix.endsWith('\t'))
    ) {
        completion = completion.slice(1)
        hasOddIndentation = true
    }

    // Experimental: Trim start of the completion to remove all trailing whitespace nonsense
    completion = completion.trimStart()

    // Detect bad completion start
    if (BAD_COMPLETION_START.test(completion)) {
        completion = completion.replace(BAD_COMPLETION_START, '')
    }

    // Strip out trailing markdown block and trim trailing whitespace
    const endBlockIndex = completion.indexOf('```')
    if (endBlockIndex !== -1) {
        completion = completion.slice(0, endBlockIndex)
    }

    if (multiline) {
        completion = truncateMultilineCompletion(completion, hasOddIndentation, prefix, nextNonEmptyLine, languageId)
    }

    // If a completed line matches the next non-empty line of the suffix 1:1, we remove
    const lines = completion.split('\n')
    const matchedLineIndex = lines.findIndex((line, index) => {
        if (index === 0) {
            line = currentLinePrefix + line
        }
        if (line.trim() !== '' && nextNonEmptyLine.trim() !== '') {
            // We need a trimEnd here because the machine likes to add trailing whitespace.
            //
            // TODO: Fix this earlier in the post process run but this needs to be careful not
            // to alter the meaning
            return line.trimEnd() === nextNonEmptyLine
        }
        return false
    })
    if (matchedLineIndex !== -1) {
        completion = lines.slice(0, matchedLineIndex).join('\n')
    }

    return completion.trimEnd()
}
