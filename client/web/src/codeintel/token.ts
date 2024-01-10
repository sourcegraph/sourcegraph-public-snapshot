/**
 * The default regex for characters allowed in an identifier. It works well for
 * C-like languages (C/C++, C#, Java, etc.) but not for languages that allow
 * punctuation characters (e.g. Ruby).
 */
const DEFAULT_IDENT_CHAR_PATTERN = /\w/

export interface ZeroBasedPosition {
    line: number
    character: number
}

/**
 * Extract the token from [start, end) if 'end' is provided, else attempt
 * to extract a token from 'start' based on common identifier characters,
 * without any language-specific heuristics.
 *
 * If a string is returned, it is guaranteed to be non-empty.
 */
export function findSearchToken(args: {
    /** The text of the current document. */
    text: string
    start: ZeroBasedPosition
    end?: ZeroBasedPosition
}): string | undefined {
    const lines = args.text.split('\n')
    if (
        args.start.line < 0 ||
        args.start.character < 0 ||
        args.start.line >= lines.length ||
        (args.end && (args.end.line !== args.start.line || args.end.character <= args.start.character))
    ) {
        return undefined
    }

    const line = lines[args.start.line]

    // In the common case, when 'Find references' is triggered through
    // the blob view, the 'end' parameter will be provided.
    if (args.end) {
        return line.slice(args.start.character, args.end.character)
    }

    // For old URLs, have a fallback where we attempt to detect identifier
    // boundaries in a best-effort fashion.

    let end = line.length
    const start = args.start.character
    for (let index = start; index < line.length; index++) {
        if (!DEFAULT_IDENT_CHAR_PATTERN.test(line[index])) {
            end = index
            break
        }
    }

    return line.slice(start, end)
}
