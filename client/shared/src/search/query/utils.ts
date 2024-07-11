import { memoize } from 'lodash'

import { type FilterType, resolveFilter } from './filters'
import type { Filter, Token } from './token'

/**
 * Returns the (string) length of the provided token.
 */
export function getTokenLength(token: Token): number {
    return token.range.end - token.range.start
}

export const resolveFilterMemoized = memoize(resolveFilter)

export function isFilterOfType(token: Token, filterType: FilterType): token is Filter {
    return token.type === 'filter' && resolveFilterMemoized(token.field.value)?.type === filterType
}
