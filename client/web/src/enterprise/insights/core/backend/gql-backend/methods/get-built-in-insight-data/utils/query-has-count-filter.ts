/**
 * Returns whether the query specifies a count. Search queries break when count is specified twice.
 */
export function queryHasCountFilter(query: string): boolean {
    return /\bcount:(\d+|all)\b/gi.test(query)
}
