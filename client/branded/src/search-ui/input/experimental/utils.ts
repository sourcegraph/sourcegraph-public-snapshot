/**
 * Takes in a file path and a desired length. Will shorten the path so that it's
 * as close to the length as possible (but doesn't gurantee it). The first two
 * and last two path segements are always visible.
 */
export function shortenPath(path: string, desiredChars: number): string {
    if (path.length <= desiredChars) {
        return path
    }

    const parts = path.split('/')
    if (parts.length < 5) {
        // We need at least 5 segements to shorten the path
        return path
    }
    const shortendPath = parts
        .slice(0, 2)
        .concat('...', ...parts.slice(-2))
        .join('/')
    return shortendPath.length < path.length ? shortendPath : path
}

/**
 * Takes a match (set of indexes) and a max length and returns a list of spans which
 * indicate the ranges that matched or didn't match. Takes an optional offset to shift
 * all positions by (i.e. the function is looking for position + offset in matches).
 * The position in the spans are inclusive.
 */
export function getSpans(matches: Set<number>, length: number, offset: number = 0): [number, number, boolean][] {
    const spans: [number, number, boolean][] = []
    let currentStart = 0
    let currentEnd = 0
    let currentMatch = false

    for (let index = 0; index < length; index++) {
        currentEnd = index
        const match = matches.has(index + offset)
        if (currentMatch !== match) {
            // close previous span
            if (currentStart !== currentEnd) {
                spans.push([currentStart, currentEnd - 1, currentMatch])
            }
            currentStart = index
            currentMatch = match
        }
    }
    // close last span
    spans.push([currentStart, currentEnd, currentMatch])

    return spans
}
