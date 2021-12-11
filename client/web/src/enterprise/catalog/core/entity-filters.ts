import * as H from 'history'
import { useCallback, useEffect, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

import { CatalogEntityType } from '@sourcegraph/shared/src/graphql/schema'

export interface CatalogEntityFilters {
    query: string
}

export interface CatalogEntityFiltersQueryFields {
    is: CatalogEntityType | undefined
}

export interface CatalogEntityFiltersProps {
    filters: CatalogEntityFilters
    filtersQueryParsed: CatalogEntityFiltersQueryFields
    onFiltersChange: (newValue: CatalogEntityFilters) => void
    onFiltersQueryFieldChange: <
        F extends keyof CatalogEntityFiltersQueryFields,
        T extends CatalogEntityFiltersQueryFields[F]
    >(
        field: F,
        newValue: T
    ) => void
}

export const useCatalogEntityFilters = (defaultQuery: string): CatalogEntityFiltersProps => {
    const history = useHistory()
    const location = useLocation()

    const filters = useMemo(() => {
        const filters = filtersFromLocation(location.search)
        return { ...filters, query: filters.query ?? defaultQuery }
    }, [defaultQuery, location.search])

    const onFiltersChange = useCallback(
        (newValue: CatalogEntityFilters) => {
            history.push({
                ...location,
                search: urlSearchParametersFromFilters(newValue, new URLSearchParams(location.search)).toString(),
            })
        },
        [history, location]
    )

    // Ensure URL reflects default query (if no ?q=, then update URL to be ?q=defaultQuery).
    useEffect(() => {
        const { query } = filtersFromLocation(location.search)
        if (query === undefined) {
            onFiltersChange({ ...filters, query: defaultQuery })
        }
    }, [defaultQuery, filters, location.search, onFiltersChange])

    const filtersQueryParsed = useMemo(() => parseFiltersQuery(filters.query), [filters.query])

    const onFiltersQueryFieldChange = useCallback<CatalogEntityFiltersProps['onFiltersQueryFieldChange']>(
        (field, newValue) => {
            onFiltersChange({ ...filters, query: queryWithField(filters.query, field, newValue) })
        },
        [filters, onFiltersChange]
    )

    return { filters, filtersQueryParsed, onFiltersChange, onFiltersQueryFieldChange }
}

function filtersFromLocation(locationSearch: H.Location['search']): Partial<CatalogEntityFilters> {
    const parameters = new URLSearchParams(locationSearch)
    return {
        query: parameters.get('q') ?? undefined,
    }
}

function urlSearchParametersFromFilters(filters: CatalogEntityFilters, base: URLSearchParams): URLSearchParams {
    const parameters = new URLSearchParams(base)

    if (filters.query !== undefined) {
        parameters.set('q', filters.query)
    } else {
        parameters.delete('q')
    }

    return parameters
}

function parseFiltersQuery(query: string): CatalogEntityFiltersQueryFields {
    const parsed: CatalogEntityFiltersQueryFields = {
        is: undefined,
    }

    for (const part of query.split(/\s+/g)) {
        for (const field of Object.keys(parsed) as (keyof CatalogEntityFiltersQueryFields)[]) {
            const prefix = `${field}:`
            if (part.startsWith(prefix)) {
                parsed[field] = part.slice(prefix.length).toUpperCase() as CatalogEntityFiltersQueryFields[typeof field]
            }
        }
    }

    return parsed
}

function queryWithField<F extends keyof CatalogEntityFiltersQueryFields, T extends CatalogEntityFiltersQueryFields[F]>(
    query: string,
    field: F,
    newValue: T
): string {
    const parts: string[] = []
    let found = false
    const prefix = `${field}:`
    for (const part of query.split(/\s+/g)) {
        if (part.startsWith(prefix)) {
            if (!found && newValue !== undefined) {
                parts.push(`${prefix}${newValue.toLowerCase()}`)
                found = true
            }
        } else {
            parts.push(part)
        }
    }
    if (!found && newValue !== undefined) {
        parts.push(`${prefix}${newValue.toLowerCase()}`)
    }
    return parts.join(' ')
}
