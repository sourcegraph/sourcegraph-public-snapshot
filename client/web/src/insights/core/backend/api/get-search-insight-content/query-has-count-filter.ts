const COUNT_PREFIX = 'count:'

/**
 * Returns whether the query specifies a count. Search queries break when count is specified twice.
 */
export function queryHasCountFilter(query: string): boolean {
    return query
        .split(' ')
        .map(part => part.trim())
        .some(part => {
            if (!part.startsWith(COUNT_PREFIX)) {
                return false
            }

            return part.slice(COUNT_PREFIX.length).length > 0
        })
}
