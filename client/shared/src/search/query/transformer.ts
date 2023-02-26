import { replaceRange } from '@sourcegraph/common'
import { SearchPatternType } from '../../graphql-operations'

import { filters, FilterType } from './filters'
import { findFilters, findFilter, FilterKind } from './query'
import { scanSearchQuery } from './scanner'
import { Filter, Token } from './token'
import { operatorExists, filterExists } from './validate'

export function appendContextFilter(
    query: string,
    searchContextSpec: string | undefined,
    enableOwnershipSearch: boolean
): string {
    return !filterExists(query, FilterType.context, false, enableOwnershipSearch) && searchContextSpec
        ? `context:${searchContextSpec} ${query}`
        : query
}

/**
 * Deletes the filter from a given query string by the filter's range.
 */
export function omitFilter(query: string, filter: Filter): string {
    const { start, end } = filter.range

    return `${query.slice(0, start).trimEnd()} ${query.slice(end).trimStart()}`.trim()
}

const succeedScan = (query: string, enableOwnershipSearch: boolean): Token[] => {
    const result = scanSearchQuery(query, false, SearchPatternType.literal, enableOwnershipSearch)
    if (result.type !== 'success') {
        throw new Error('Internal error: invariant broken: succeedScan callers must be called with a valid query')
    }
    return result.term
}

/**
 * Updates the first filter with the given value if it exists.
 * Appends a single filter at the top level of the query if it does not exist.
 * This function expects a valid query; if it is invalid it throws.
 */
export const updateFilter = (query: string, field: string, value: string, enableOwnershipSearch: boolean): string => {
    const filters = findFilters(succeedScan(query, enableOwnershipSearch), field)
    return filters.length > 0
        ? replaceRange(query, filters[0].range, `${field}:${value}`).trim()
        : `${query} ${field}:${value}`
}

/**
 * Updates all filters with the given value if they exist.
 * Appends a single filter at the top level of the query if none exist.
 * This function expects a valid query; if it is invalid it throws.
 */
export const updateFilters = (query: string, field: string, value: string, enableOwnershipSearch: boolean): string => {
    const filters = findFilters(succeedScan(query, enableOwnershipSearch), field)
    let modified = false
    for (const filter of filters.reverse()) {
        query = replaceRange(query, filter.range, `${field}:${value}`)
        modified = true
    }
    if (modified) {
        return query.trim()
    }
    return `${query} ${field}:${value}`
}

/**
 * Appends the provided filter.
 */
export const appendFilter = (query: string, field: string, value: string): string => {
    const trimmedQuery = query.trim()
    const filter = `${field}:${value}`
    return trimmedQuery.length === 0 ? filter : `${query.trimEnd()} ${filter}`
}

/**
 * Removes certain filters from a given query for privacy purposes, so query can be logged in telemtry.
 */
export const sanitizeQueryForTelemetry = (query: string, enableOwnershipSearch: boolean): string => {
    const redactedValue = '[REDACTED]'
    const filterToRedact = [
        FilterType.repo,
        FilterType.file,
        FilterType.rev,
        FilterType.repohasfile,
        FilterType.context,
        FilterType.message,
    ]

    let newQuery = query

    for (const filter of filterToRedact) {
        if (filterExists(query, filter, false, enableOwnershipSearch)) {
            newQuery = updateFilters(newQuery, filter, redactedValue, enableOwnershipSearch)
        }
        if (filterExists(query, filter, true, enableOwnershipSearch)) {
            newQuery = updateFilters(newQuery, `-${filter}`, redactedValue, enableOwnershipSearch)
        }
        const alias = filters(enableOwnershipSearch)[filter].alias
        if (alias) {
            if (filterExists(query, alias, false, enableOwnershipSearch)) {
                newQuery = updateFilters(newQuery, alias, redactedValue, enableOwnershipSearch)
            }
            if (filterExists(query, alias, true, enableOwnershipSearch)) {
                newQuery = updateFilters(newQuery, `-${alias}`, redactedValue, enableOwnershipSearch)
            }
        }
    }

    return newQuery
}

/**
 * Wraps a query in parenthesis if a global search context filter exists.
 * Example: context:ctx a or b -> context:ctx (a or b)
 */
export function parenthesizeQueryWithGlobalContext(query: string, enableOwnershipSearch: boolean): string {
    if (!operatorExists(query, enableOwnershipSearch)) {
        // no need to parenthesize a flat, atomic query.
        return query
    }
    const globalContextFilter = findFilter(query, FilterType.context, FilterKind.Global, enableOwnershipSearch)
    if (!globalContextFilter) {
        // don't parenthesize a query that contains `context` subexpressions already.
        return query
    }
    const searchContextSpec = globalContextFilter.value?.value || ''
    const queryWithOmittedContext = omitFilter(query, globalContextFilter)
    return appendContextFilter(`(${queryWithOmittedContext})`, searchContextSpec, enableOwnershipSearch)
}
