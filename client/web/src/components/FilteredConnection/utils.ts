import { GraphQLError } from 'graphql'
import type { Location } from 'history'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'
import { hasProperty } from '@sourcegraph/shared/src/util/types'

import { Connection } from './ConnectionType'
import { QUERY_KEY } from './constants'
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

export const hasNextPage = (connection: Connection<unknown>): boolean =>
    connection.pageInfo
        ? connection.pageInfo.hasNextPage
        : typeof connection.totalCount === 'number' && connection.nodes.length < connection.totalCount

export interface GetUrlQueryParameters {
    first?: number
    defaultFirst?: number
    query?: string
    values?: Map<string, FilteredConnectionFilterValue>
    filters?: FilteredConnectionFilter[]
    visible?: number
    location: Location
}

export const getUrlQuery = ({
    first,
    defaultFirst,
    query,
    values,
    visible,
    filters,
    location,
}: GetUrlQueryParameters): string => {
    const searchParameters = new URLSearchParams(location.search)
    if (query) {
        searchParameters.set(QUERY_KEY, query)
    }

    if (first !== defaultFirst) {
        searchParameters.set('first', String(first))
    }

    if (values && filters) {
        for (const filter of filters) {
            if (values === undefined) {
                continue
            }
            const value = values.get(filter.id)
            if (value === undefined) {
                continue
            }
            if (value !== filter.values[0]) {
                searchParameters.set(filter.id, value.value)
            } else {
                searchParameters.delete(filter.id)
            }
        }
    }

    if (visible !== 0 && visible !== first) {
        searchParameters.set('visible', String(visible))
    }

    return searchParameters.toString()
}

interface asGraphQLResultParameters<TResult> {
    data: TResult | null
    errors: readonly GraphQLError[]
}

export const asGraphQLResult = <T>({ data, errors }: asGraphQLResultParameters<T>): GraphQLResult<T> => {
    if (!data) {
        return { data: undefined, errors }
    }
    return {
        data,
        errors: undefined,
    }
}
