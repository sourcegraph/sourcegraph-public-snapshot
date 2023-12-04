import { zip } from 'lodash'

type LineKind = 'prose' | 'code'

/** Regular express to match a code fence */
const commentPattern = /```/

/** Regular express to match a closing HTML tag */
const htmlTagPattern = /<\//

/** Regular express to match a markdown list */
const markdownListPattern = /^(1\.|- |\* )/m

/** The patterns that indicate text is already formatted */
const excludelistPatterns = [commentPattern, htmlTagPattern, markdownListPattern]

/**
 * Wrap lines that appear to be code blocks in ```fences``` tagged with the given language
 * identifier. This function will not try to format something that appears to be pre-formatted.
 *
 * @param languageID The language identifier.
 * @param docstring The raw docstring.
 */
export function wrapIndentationInCodeBlocks(languageID: string, docstring: string): string {
    if (excludelistPatterns.some(pattern => pattern.test(docstring))) {
        // Already formatted
        return docstring
    }

    // Determine categories of each line
    const lines = categorize(docstring.split('\n'))

    // Create a zipped sequence of categorized lines of the form
    //    [[line1, line2], [line2, line3], ..., [line-n, undef]]

    const pairs = zip(lines, lines.slice(1)) as [{ line: string; kind: LineKind }, { kind: LineKind } | undefined][]

    return pairs
        .flatMap(([{ line, kind }, { kind: nextKind } = { kind: undefined }]) => {
            if (kind === 'prose' && nextKind === 'code') {
                // going from prose to code, start a code block
                return [line, '```' + languageID]
            }
            if (kind === 'code' && nextKind !== 'code') {
                // going from code to non-code, close the last code block
                return [line, '```']
            }

            return [line]
        })
        .join('\n')
}

const reducer = (
    last: LineKind | undefined,
    line: { line: string; kind: LineKind | undefined }
): LineKind | undefined => {
    line.kind = line.kind || (last === 'prose' ? 'prose' : undefined)
    return line.kind
}

/**
 * Decorate each line of text with its kind (prose or code) depending on the
 * context in which it occurs.
 *
 * @param lines An array of lines.
 */
function categorize(lines: string[]): { line: string; kind: LineKind | undefined }[] {
    // Find an initial category for lines that don't need additional context
    const categorizedLines = lines.map(line => ({ line, kind: kindOf(line) }))

    // Each empty or whitespace-only line is code if it's surrounded by code
    // and prose otherwise. Mark every undefined line that's "reachable" by
    // prose as prose.

    const reversedCategorizedLines = categorizedLines.slice().reverse()
    categorizedLines.reduce(reducer, 'prose' as LineKind | undefined)
    reversedCategorizedLines.reduce(reducer, 'prose' as LineKind | undefined)

    // All remaining undefined lines are code
    for (const line of categorizedLines) {
        line.kind = line.kind || 'code'
    }

    return categorizedLines
}

/**
 * Determine the type of line:
 *
 * - prose lines start with no-whitespace at the beginning of the line
 * - code lines start with two spaces or a `>` followed by non-whitespace
 * - empty lines or whitespace-only lines are of an unknown type
 *
 * @param line The line content.
 */
function kindOf(line: string): LineKind | undefined {
    if (/^\S/.test(line)) {
        return 'prose'
    }
    if (/^( {2}|>).*\S/.test(line)) {
        return 'code'
    }
    return undefined
}
