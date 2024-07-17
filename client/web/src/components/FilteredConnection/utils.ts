import type { GraphQLError } from 'graphql'
import type { Location } from 'history'

import { hasProperty } from '@sourcegraph/common'
import type { GraphQLResult } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import type { Connection } from './ConnectionType'
import { QUERY_KEY } from './constants'
import type { Filter, FilterValues } from './FilterControl'
import type { PaginatedConnectionQueryArguments } from './hooks/usePageSwitcherPagination'

/** Checks if the passed value satisfies the GraphQL Node interface */
export function hasID(value: unknown): value is { id: Scalars['ID'] } {
    return typeof value === 'object' && value !== null && hasProperty('id')(value) && typeof value.id === 'string'
}

export function hasDisplayName(value: unknown): value is { displayName: Scalars['String'] } {
    return (
        typeof value === 'object' &&
        value !== null &&
        hasProperty('displayName')(value) &&
        typeof value.displayName === 'string'
    )
}

export function getFilterFromURL<K extends string>(
    searchParameters: URLSearchParams,
    filters: Filter<K>[] | undefined
): FilterValues<K> {
    const values: FilterValues<K> = {}
    if (filters === undefined) {
        return values
    }
    for (const filter of filters) {
        const urlValue = searchParameters.get(filter.id)
        if (urlValue !== null) {
            const value = filter.options.find(opt => opt.value === urlValue)
            if (value !== undefined) {
                values[filter.id] = value.value
                continue
            }
        }

        // couldn't find a value, add default
        const defaultOption = filter.options.at(0)
        if (defaultOption !== undefined) {
            values[filter.id] = defaultOption.value
        }
    }
    return values
}

export function parseQueryInt(searchParameters: URLSearchParams, name: string): number | null {
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

/**
 * Determine if a connection has a next page.
 * Provides fallback logic to support queries where `hasNextPage` is undefined.
 */
export const hasNextPage = (connection: Connection<unknown>): boolean =>
    connection.pageInfo
        ? connection.pageInfo.hasNextPage
        : typeof connection.totalCount === 'number' && connection.nodes.length < connection.totalCount

/**
 * Determines the URL search parameters for a connection. All of the parameters that may be used in
 * a filtered connection are handled here: search query, filters (where the URL querystring params
 * differ from the actual args that are passed as GraphQL variables), connection pagination params
 * like `first` and `after`, etc.
 */
export function urlSearchParamsForFilteredConnection({
    pagination,
    pageSize,
    query,
    filterValues,
    filters,
    search,
}: {
    pagination?: PaginatedConnectionQueryArguments
    pageSize?: number
    query?: string
    filterValues?: FilterValues
    filters?: Filter[]
    search: Location['search']
}): URLSearchParams {
    const params = new URLSearchParams(search)

    setOrDeleteSearchParam(params, QUERY_KEY, query)

    if (pagination) {
        // Omit `first` or `last` if their value is the default page size and if they are implicit
        // because it's just noise in the URL.
        const firstIfNonDefault =
            pageSize !== undefined && pagination.first === pageSize && !pagination.before && !pagination.last
                ? null
                : pagination.first
        const lastIfNonDefault =
            pageSize !== undefined &&
            pagination.last === pageSize &&
            pagination.before &&
            !pagination.after &&
            !pagination.first
                ? null
                : pagination.last
        setOrDeleteSearchParam(params, 'first', firstIfNonDefault)
        setOrDeleteSearchParam(params, 'last', lastIfNonDefault)
        setOrDeleteSearchParam(params, 'before', pagination.before)
        setOrDeleteSearchParam(params, 'after', pagination.after)
    }

    if (filterValues && filters) {
        for (const filter of filters) {
            const value = filterValues[filter.id]
            const defaultOption = filter.options.at(0)
            if (value !== undefined && value !== null && (!defaultOption || value !== defaultOption.value)) {
                params.set(filter.id, value)
            } else {
                params.delete(filter.id)
            }
        }
    }

    return params
}

function setOrDeleteSearchParam(
    params: URLSearchParams,
    name: string,
    value: string | number | null | undefined
): void {
    if (value !== null && value !== undefined && value !== '' && value !== 0) {
        params.set(name, value.toString())
    } else {
        params.delete(name)
    }
}

/**
 * Map non-conforming GraphQL responses to a GraphQLResult.
 */
export function asGraphQLResult<T>({ data, errors }: { data?: T; errors: readonly GraphQLError[] }): GraphQLResult<T> {
    if (!data) {
        return { data: null, errors }
    }
    return {
        data,
        errors: undefined,
    }
}
