import { Filter, scanSearchQuery } from './scanner'

/**
 * Returns a global filter for a field in a query, if any. A filter is
 * global iff (1) it is specified once and (2) it is at the top-level of a query.
 * For example, `case:yes` is not global in the following queries:
 *
 * `(case:yes some subexpression) case:no multiple cases`
 * `(case:yes not at top level; inside a parentheses of a grouped expression)`
 *
 * @param query the query string
 * @param field the field of the filter to find
 */
export const findGlobalFilter = (query: string, field: string): Filter | undefined => {
    const result = scanSearchQuery(query)
    let filter: Filter | undefined
    if (result.type === 'success') {
        let depth = 0
        let seenField = false
        for (const token of result.term) {
            if (token.type === 'openingParen') {
                depth = depth + 1
            }
            if (token.type === 'closingParen') {
                depth = depth - 1
            }
            if (token.type === 'filter' && token.field.value.toLowerCase() === field) {
                if (seenField) {
                    // More than one of this field.
                    return undefined
                }
                if (depth > 0) {
                    // Inside a grouped expression.
                    return undefined
                }
                filter = token
                seenField = true
                continue
            }
        }
    }
    return filter
}
