import { replaceRange } from '../../util/strings'

import { FilterType } from './filters'
import { scanSearchQuery } from './scanner'
import { Filter, Token } from './token'
import { filterExists, findFilters } from './validate'

export function appendContextFilter(query: string, searchContextSpec: string | undefined): string {
    return !filterExists(query, FilterType.context) && searchContextSpec
        ? `context:${searchContextSpec} ${query}`
        : query
}

export function omitFilter(query: string, filter: Filter): string {
    return replaceRange(query, filter.range).trimStart()
}

const succeedScan = (query: string): Token[] => {
    const result = scanSearchQuery(query)
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
export const updateFilter = (query: string, field: string, value: string): string => {
    const filters = findFilters(succeedScan(query), field)
    return filters.length > 0
        ? replaceRange(query, filters[0].range, `${field}:${value}`).trim()
        : `${query} ${field}:${value}`
}

/**
 * Updates all filters with the given value if they exist.
 * Appends a single filter at the top level of the query if none exist.
 * This function expects a valid query; if it is invalid it throws.
 */
export const updateFilters = (query: string, field: string, value: string): string => {
    const filters = findFilters(succeedScan(query), field)
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
