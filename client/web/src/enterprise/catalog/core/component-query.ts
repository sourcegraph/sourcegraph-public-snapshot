import * as H from 'history'
import { useCallback, useEffect, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

import { ComponentKind } from '@sourcegraph/shared/src/schema'

export interface ComponentFilters {
    query: string
}

export interface ComponentFiltersQueryFields {
    is: ComponentKind | undefined
    tag: string | undefined
}

export interface ComponentFiltersProps {
    filters: ComponentFilters
    filtersQueryParsed: ComponentFiltersQueryFields
    onFiltersChange: (newValue: ComponentFilters) => void
    onFiltersQueryFieldChange: <F extends keyof ComponentFiltersQueryFields, T extends ComponentFiltersQueryFields[F]>(
        field: F,
        newValue: T
    ) => void
}

export const useComponentFilters = (defaultQuery: string): ComponentFiltersProps => {
    const history = useHistory()
    const location = useLocation()

    const filters = useMemo(() => {
        const filters = filtersFromLocation(location.search)
        return { ...filters, query: filters.query ?? defaultQuery }
    }, [defaultQuery, location.search])

    const onFiltersChange = useCallback(
        (newValue: ComponentFilters) => {
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
        if (query === undefined && defaultQuery !== '') {
            onFiltersChange({ ...filters, query: defaultQuery })
        }
    }, [defaultQuery, filters, location.search, onFiltersChange])

    const filtersQueryParsed = useMemo(() => parseFiltersQuery(filters.query), [filters.query])

    const onFiltersQueryFieldChange = useCallback<ComponentFiltersProps['onFiltersQueryFieldChange']>(
        (field, newValue) => {
            onFiltersChange({ ...filters, query: queryWithField(filters.query, field, newValue) })
        },
        [filters, onFiltersChange]
    )

    return { filters, filtersQueryParsed, onFiltersChange, onFiltersQueryFieldChange }
}

function filtersFromLocation(locationSearch: H.Location['search']): Partial<ComponentFilters> {
    const parameters = new URLSearchParams(locationSearch)
    return {
        query: parameters.get('q') ?? undefined,
    }
}

function urlSearchParametersFromFilters(filters: ComponentFilters, base: URLSearchParams): URLSearchParams {
    const parameters = new URLSearchParams(base)

    if (filters.query !== '') {
        parameters.set('q', filters.query)
    } else {
        parameters.delete('q')
    }

    return parameters
}

function parseFiltersQuery(query: string): ComponentFiltersQueryFields {
    const parsed: ComponentFiltersQueryFields = {
        is: undefined,
        tag: undefined,
    }

    for (const part of query.split(/\s+/g)) {
        for (const field of Object.keys(parsed) as (keyof ComponentFiltersQueryFields)[]) {
            const prefix = `${field}:`
            if (part.startsWith(prefix)) {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                parsed[field] = (part.slice(prefix.length).toUpperCase() as unknown) as any
            }
        }
    }

    return parsed
}

function queryWithField<F extends keyof ComponentFiltersQueryFields, T extends ComponentFiltersQueryFields[F]>(
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
