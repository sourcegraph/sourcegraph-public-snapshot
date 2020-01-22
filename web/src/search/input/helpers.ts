import { FiltersToTypeAndValue, FilterTypes } from '../../../../shared/src/search/interactive/util'
import { parseSearchQuery } from '../../../../shared/src/search/parser/parser'
import { uniqueId } from 'lodash'

/**
 * Creates the raw string representation of the filters currently in the query in interactive mode.
 *
 * @param filtersInQuery the map representing the filters currently in an interactive mode query.
 */
export function generateFiltersQuery(filtersInQuery: FiltersToTypeAndValue): string {
    const fieldKeys = Object.keys(filtersInQuery)
    const individualTokens: string[] = []
    fieldKeys
        .filter(key => filtersInQuery[key].value.trim().length > 0)
        .map(key => individualTokens.push(`${filtersInQuery[key].type}:${filtersInQuery[key].value}`))

    return individualTokens.join(' ')
}

export function convertPlainTextToInteractiveQuery(
    query: string,
): { filtersInQuery: FiltersToTypeAndValue; navbarQuery: string } {
    const parsedQuery = parseSearchQuery(query)

    const newFiltersInQuery: FiltersToTypeAndValue = {}
    let newNavbarQuery = ''

    if (parsedQuery.type === 'success') {
        for (const member of parsedQuery.token.members) {
            if (member.token.type === 'filter' && member.token.filterValue) {
                newFiltersInQuery[uniqueId(member.token.filterType.token.value)] = {
                    type: member.token.filterType.token.value as FilterTypes,
                    value: query.substring(member.token.filterValue.range.start, member.token.filterValue.range.end),
                    editable: false,
                }
            } else if (member.token.type === 'literal' || member.token.type === 'quoted') {
                newNavbarQuery += ` ${query.substring(member.range.start, member.range.end)}`
            }
        }
    }

    return { filtersInQuery: newFiltersInQuery, navbarQuery: newNavbarQuery }
}
