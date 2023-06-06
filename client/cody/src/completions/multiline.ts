import * as vscode from 'vscode'

export function detectMultilineMode(
    prefix: string,
    prevNonEmptyLine: string,
    sameLinePrefix: string,
    sameLineSuffix: string,
    languageId: string
): null | 'block' {
    const config = getLanguageConfig(languageId)

    if (
        config &&
        // Only trigger multiline suggestions for empty lines
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

export function truncateMultilineCompletion(
    completion: string,
    hasOddIndentation: boolean,
    prefix: string,
    nextNonEmptyLine: string,
    languageId: string
): string {
    const config = getLanguageConfig(languageId)
    if (!config) {
        return completion
    }

    const lines = completion.split('\n')

    // We use a whitespace counting approach to finding the end of the completion. To find
    // an end, we look for the first line that is below the start scope of the completion (
    // calculated by the number of leading spaces or tabs)
    const prefixLastNewline = prefix.lastIndexOf('\n')
    const prefixIndentationWithFirstCompletionLine = prefix.slice(prefixLastNewline + 1) + completion[0]
    const startIndent = indentation(prefixIndentationWithFirstCompletionLine)

    // Normalize responses that start with a newline followed by the exact indentation of
    // the first line.
    if (lines.length > 1 && lines[0] === '' && indentation(lines[1]) === startIndent) {
        lines.shift()
        lines[0] = lines[0].trimStart()
    }

    // If odd indentation is detected (i.e Claude adds a space to every line),
    // we fix it for the whole multiline block first.
    //
    // We can skip the first line as it was already corrected above
    if (hasOddIndentation) {
        for (let i = 1; i < lines.length; i++) {
            if (indentation(lines[i]) >= startIndent) {
                lines[i] = lines[i].replace(/^(\t)* /, '$1')
            }
        }
    }

    // Only include a closing line (e.g. `}`) if the block is empty yet. We detect this by
    // looking at the indentation of the next non-empty line.
    const includeClosingLine = indentation(nextNonEmptyLine) < startIndent

    let cutOffIndex = lines.length
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i]

        if (i === 0 || line === '' || config.blockElseTest.test(line)) {
            continue
        }

        if (indentation(line) < startIndent) {
            // When we find the first block below the start indentation, only include it if
            // it is an end block
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

/**
 * Counts space or tabs in the beginning of a line.
 *
 * Since Cody can sometimes respond in a mix of tab and spaces, this function
 * normalizes the whitespace first using the currently enabled tabSize option.
 */
function indentation(line: string): number {
    const tabSize = vscode.window.activeTextEditor
        ? // tabSize is always resolved to a number when accessing the property
          (vscode.window.activeTextEditor.options.tabSize as number)
        : 2

    const regex = line.match(/^[\t ]*/)
    if (regex) {
        const whitespace = regex[0]
        return [...whitespace].reduce((p, c) => p + (c === '\t' ? tabSize : 1), 0)
    }
    return 0
}

interface LanguageConfig {
    blockStart: string
    blockElseTest: RegExp
    blockEnd: string | null
}
function getLanguageConfig(languageId: string): LanguageConfig | null {
    switch (languageId) {
        case 'typescript':
        case 'javascript':
        case 'go':
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
