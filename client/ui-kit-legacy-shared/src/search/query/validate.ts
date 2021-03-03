import { FilterType } from './filters'
import { scanSearchQuery } from './scanner'
import { Filter } from './token'

export enum FilterKind {
    Global = 'Global',
    Subexpression = 'Subexpression',
}

/**
 * Returns the first filter for a field in a query, if any. A FilterKind
 * specifies what kind of filter to look for.
 *
 * A Global filter is found iff (1) it is specified once and (2) it is at
 * the top-level of a query.
 *
 * A Subexpression filter is found if a non-global filter exists. For
 * example, `case:yes` is not global, but are part of subexpressions in
 * the following queries:
 *
 * `(case:yes some subexpression) case:no multiple cases`
 * `(case:yes not at top level; inside a parentheses of a grouped expression)`
 *
 * @param query the query string
 * @param field the field of the filter to find
 * @param kind the kind of filter to find
 */
export const findFilter = (query: string, field: string, kind: FilterKind): Filter | undefined => {
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
                    return kind === FilterKind.Subexpression ? token : undefined
                }
                if (depth > 0) {
                    // Inside a grouped expression.
                    return kind === FilterKind.Subexpression ? token : undefined
                }
                filter = token
                seenField = true
            }
        }
    }
    return kind === FilterKind.Global ? filter : undefined
}

export function isContextFilterInQuery(query: string): boolean {
    const scannedQuery = scanSearchQuery(query)
    return (
        scannedQuery.type === 'success' &&
        scannedQuery.term.some(
            token => token.type === 'filter' && token.field.value.toLowerCase() === FilterType.context
        )
    )
}
