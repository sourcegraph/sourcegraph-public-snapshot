import { FilterType, resolveFilter } from './filters'
import type { Filter, Token } from './token'

/**
 * Returns the (string) length of the provided token.
 */
export function getTokenLength(token: Token): number {
    return token.range.end - token.range.start
}

export function isFilterOfType(filter: Filter, type: FilterType): boolean {
    return resolveFilter(filter.field.value)?.type === type
}
