import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { hasProperty } from '@sourcegraph/shared/src/util/types'

import type { FilteredConnectionFilter, FilteredConnectionFilterValue } from './FilterControl'

/** Checks if the passed value satisfies the GraphQL Node interface */
export const hasID = (value: unknown): value is { id: Scalars['ID'] } =>
    typeof value === 'object' && value !== null && hasProperty('id')(value) && typeof value.id === 'string'

export const getFilterFromURL = (
    searchParameters: URLSearchParams,
    filters: FilteredConnectionFilter[] | undefined
): Map<string, FilteredConnectionFilterValue> => {
    const values: Map<string, FilteredConnectionFilterValue> = new Map<string, FilteredConnectionFilterValue>()

    if (filters === undefined || filters.length === 0) {
        return values
    }
    for (const filter of filters) {
        const urlValue = searchParameters.get(filter.id)
        if (urlValue !== null) {
            const value = filter.values.find(value => value.value === urlValue)
            if (value !== undefined) {
                values.set(filter.id, value)
                continue
            }
        }
        // couldn't find a value, add default
        values.set(filter.id, filter.values[0])
    }
    return values
}

export const parseQueryInt = (searchParameters: URLSearchParams, name: string): number | null => {
    const valueString = searchParameters.get(name)
    if (valueString === null) {
        return null
    }
    const valueNumber = parseInt(valueString, 10)
    if (valueNumber > 0) {
        return valueNumber
    }
    return null
}
