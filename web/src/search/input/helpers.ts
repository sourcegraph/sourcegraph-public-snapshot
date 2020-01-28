import { FiltersToTypeAndValue, FilterTypes } from '../../../../shared/src/search/interactive/util'
import { parseSearchQuery } from '../../../../shared/src/search/parser/parser'
import { uniqueId } from 'lodash'

export function convertPlainTextToInteractiveQuery(
    query: string
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
                newNavbarQuery = [newNavbarQuery, query.substring(member.range.start, member.range.end)]
                    .filter(query => query.length > 0)
                    .join(' ')
            }
        }
    }

    return { filtersInQuery: newFiltersInQuery, navbarQuery: newNavbarQuery }
}
