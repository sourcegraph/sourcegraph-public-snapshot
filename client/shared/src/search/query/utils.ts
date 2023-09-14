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

export function isFilterOfType(filter: Filter, filterType: FilterType): boolean {
    return resolveFilterMemoized(filter.field.value)?.type === filterType
}
