import {
    FiltersToTypeAndValue,
    FilterType,
    isNegatedFilter,
    resolveNegatedFilter,
} from '../../../../shared/src/search/interactive/util'
import { scanSearchQuery } from '../../../../shared/src/search/parser/scanner'
import { uniqueId } from 'lodash'
import { validateFilter, isSingularFilter } from '../../../../shared/src/search/parser/filters'

/**
 * Converts a plain text query into a an object containing the two components
 * of an interactive mode query:
 * - navbarQuery: any non-filter values in the query, which appears in the main query input
 * - filtersInQuery: an object containing key-value pairs of filters and their values
 *
 * @param query a plain text query.
 */
export function convertPlainTextToInteractiveQuery(
    query: string
): { filtersInQuery: FiltersToTypeAndValue; navbarQuery: string } {
    const scannedQuery = scanSearchQuery(query)

    const newFiltersInQuery: FiltersToTypeAndValue = {}
    let newNavbarQuery = ''

    if (scannedQuery.type === 'success') {
        for (const token of scannedQuery.term) {
            if (token.type === 'filter' && token.value && validateFilter(token.field.value, token.value).valid) {
                const filterType = token.field.value as FilterType
                newFiltersInQuery[isSingularFilter(filterType) ? filterType : uniqueId(filterType)] = {
                    type: isNegatedFilter(filterType) ? resolveNegatedFilter(filterType) : filterType,
                    value: query.slice(token.value.range.start, token.value.range.end),
                    editable: false,
                    negated: isNegatedFilter(filterType),
                }
            } else if (
                token.type !== 'filter' ||
                (token.type === 'filter' && !validateFilter(token.field.value, token.value).valid)
            ) {
                newNavbarQuery = [newNavbarQuery, query.slice(token.range.start, token.range.end)]
                    .filter(query => query.length > 0)
                    .join('')
            }
        }
    }

    return { filtersInQuery: newFiltersInQuery, navbarQuery: newNavbarQuery.trim() }
}
