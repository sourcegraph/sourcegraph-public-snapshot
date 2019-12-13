import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'

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
