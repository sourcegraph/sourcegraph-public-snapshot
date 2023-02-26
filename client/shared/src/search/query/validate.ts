import { SearchPatternType } from '../../graphql-operations'

import {
    AliasedFilterType,
    filters,
    FilterType,
    isNegatedFilter,
    resolveFieldAlias,
    resolveNegatedFilter,
} from './filters'
import { scanSearchQuery } from './scanner'
import { Filter, Token } from './token'

/** Returns true if the query contains operators. */
export const operatorExists = (query: string, enableOwnershipSearch: boolean): boolean => {
    const result = scanSearchQuery(query, false, SearchPatternType.literal, enableOwnershipSearch)
    return result.type === 'success' && result.term.some(term => term.type === 'keyword')
}

/**
 * Returns true if the query contains a pattern.
 */
export const containsLiteralOrPattern = (
    query: string,
    searchPatternType: SearchPatternType | undefined,
    enableOwnershipSearch: boolean
): boolean => {
    const result = scanSearchQuery(query, undefined, searchPatternType, enableOwnershipSearch)
    return result.type === 'success' && result.term.some(term => term.type === 'literal' || term.type === 'pattern')
}

/**
 * Type guard for repo: filter token.
 *
 * @param token - query parsed lexical token
 */
export const isRepoFilter = (token: Token): token is Filter =>
    token.type === 'filter' &&
    (token.field.value === FilterType.repo || token.field.value === filters(true)[FilterType.repo].alias)

/**
 * Type guard for arbitrary filter type. Also handles aliased and negated filters.
 *
 * @param token - query parsed lexical token
 */
export const isFilterType = (token: Token, filterType: FilterType): token is Filter =>
    token.type === 'filter' &&
    (token.field.value === filterType ||
        resolveFieldAlias(token.field.value) === filterType ||
        (isNegatedFilter(token.field.value) && resolveNegatedFilter(token.field.value) === filterType))

export function filterExists(
    query: string,
    filter: FilterType | keyof typeof AliasedFilterType,
    negated: boolean = false,
    enableOwnershipSearch: boolean
): boolean {
    const scannedQuery = scanSearchQuery(query, false, SearchPatternType.literal, enableOwnershipSearch)
    return (
        scannedQuery.type === 'success' &&
        scannedQuery.term.some(
            token => token.type === 'filter' && token.field.value.toLowerCase() === `${negated ? '-' : ''}${filter}`
        )
    )
}
