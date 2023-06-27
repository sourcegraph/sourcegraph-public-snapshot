import { Completion, CompletionsTextEditor } from '.'
import { truncateMultilineCompletion } from './multiline'

const BAD_COMPLETION_START = /^(\p{Emoji_Presentation}|\u{200B}|\+ |- |. )+(\s)+/u

export function postProcess({
    textEditor,
    prefix,
    suffix,
    languageId,
    multiline,
    completion,
}: {
    textEditor: CompletionsTextEditor
    prefix: string
    suffix: string
    languageId: string
    multiline: boolean
    completion: Completion
}): Completion {
    let content = completion.content

    // Extract a few common parts for the processing
    const currentLinePrefix = prefix.slice(prefix.lastIndexOf('\n') + 1)
    const firstNlInSuffix = suffix.indexOf('\n') + 1
    const nextNonEmptyLine =
        suffix
            .slice(firstNlInSuffix)
            .split('\n')
            .find(line => line.trim().length > 0) ?? ''

    // Sometimes Claude emits a single space in the completion. We call this an "odd indentation"
    // completion and try to fix the response.
    let hasOddIndentation = false
    if (
        content.length > 0 &&
        /^ [^ ]/.test(content) &&
        prefix.length > 0 &&
        (prefix.endsWith(' ') || prefix.endsWith('\t'))
    ) {
        content = content.slice(1)
        hasOddIndentation = true
    }

    // Experimental: Trim start of the completion to remove all trailing whitespace nonsense
    content = content.trimStart()

    // Detect bad completion start
    if (BAD_COMPLETION_START.test(content)) {
        content = content.replace(BAD_COMPLETION_START, '')
    }

    // Strip out trailing markdown block and trim trailing whitespace
    const endBlockIndex = content.indexOf('```')
    if (endBlockIndex !== -1) {
        content = content.slice(0, endBlockIndex)
    }

    if (multiline) {
        content = truncateMultilineCompletion(
            textEditor,
            content,
            hasOddIndentation,
            prefix,
            nextNonEmptyLine,
            languageId
        )
    } else if (content.includes('\n')) {
        content = content.slice(0, content.indexOf('\n'))
    }

    // If a completed line matches the next non-empty line of the suffix 1:1, we remove
    const lines = content.split('\n')
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
        content = lines.slice(0, matchedLineIndex).join('\n')
    }

    return {
        ...completion,
        content: content.trimEnd(),
    }
}
