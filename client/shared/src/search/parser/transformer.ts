import { scanSearchQuery, toQueryString } from './scanner'

export interface TransformSuccess {
    type: 'success'
    value: string
}

export interface TransformError {
    type: 'error'
    reason: string
}

export type TransformResult = TransformError | TransformSuccess

/**
 * Erases a single parameter like `case:yes` for field `case` at the toplevel of a query. Fails
 * with {@link TransformError} otherwise.
 *
 * @param query The input query string
 * @param field The parameter field indicating the parameter to erase.
 */
export const eraseTopLevelParameter = (query: string, field: string): TransformResult => {
    const tokens = scanSearchQuery(query) // pattern type effectively doesn't matter when erasing parameters.
    const newTokens = []
    if (tokens.type === 'success') {
        let balanced = 0
        let depth = 0
        let seenField = false
        for (const token of tokens.term) {
            if (token.type === 'openingParen') {
                balanced = balanced + 1
                depth = depth + 1
            }
            if (token.type === 'closingParen') {
                balanced = balanced - 1
            }
            if (token.type === 'filter' && token.filterType.value === field) {
                if (seenField) {
                    return {
                        type: 'error',
                        reason: `can only transform query with one ${field}`,
                    }
                }
                if (depth > 0) {
                    return {
                        type: 'error',
                        reason:
                            'can only transform parameter at the toplevel, this one occurs in a grouped subexprsesion',
                    }
                }
                seenField = true
                continue
            }
            newTokens.push(token)
        }
    }
    return {
        type: 'success',
        value: toQueryString(newTokens),
    }
}
