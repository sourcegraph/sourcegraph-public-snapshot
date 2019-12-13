import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'

/**
 * Creates the raw string representation of the filters currently in the query in interactive mode.
 *
 * @param filtersInQuery the map representing the filters currently in an interactive mode query.
 */
export const generateFiltersQuery = (filtersInQuery: FiltersToTypeAndValue): string => {
    Object.keys(filtersInQuery)
        .filter(key => filtersInQuery[key].value.trim().length > 0)
        .reduce((tokens, key) => tokens.concat(`${filtersInQuery[key].type}:${filtersInQuery[key].value}`))
}
