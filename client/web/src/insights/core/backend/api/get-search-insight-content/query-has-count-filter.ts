/**
 * Returns whether the query specifies a count. Search queries break when count is specified twice.
 */
export function queryHasCountFilter(query: string): boolean {
    return /(?<!\s["'])count:(\s*)\d+\b(?!(\s*)["'])/gi.test(query)
}
