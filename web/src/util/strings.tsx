/**
 * Returns the sum of the number of matches of the patterns in the string.
 * @param patterns Patterns to match in the string.
 */
export function count(str: string, ...patterns: RegExp[]): number {
    let n = 0
    for (const p of patterns) {
        if (!p.global) {
            throw new Error('expected RegExp to be global (or else count is inaccurate)')
        }
        const m = str.match(p)
        if (m) {
            n += m.length
        }
    }
    return n
}
