import * as H from 'history'
import { useCallback, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

export interface CatalogComponentFilters {
    query?: string
    owner?: string
    system?: string
    tags?: string[]
}

export interface CatalogComponentFiltersProps {
    filters: CatalogComponentFilters
    onFiltersChange: (newValue: CatalogComponentFilters) => void
}

export const useCatalogComponentFilters = (): CatalogComponentFiltersProps => {
    const history = useHistory()
    const location = useLocation()

    const filters = useMemo(() => catalogComponentFiltersFromLocation(location.search), [location.search])
    const onFiltersChange = useCallback(
        (newValue: CatalogComponentFilters) => {
            history.push({
                ...location,
                search: urlSearchParamsFromCatalogComponentFilters(
                    newValue,
                    new URLSearchParams(location.search)
                ).toString(),
            })
        },
        [history, location]
    )

    return { filters, onFiltersChange }
}

function catalogComponentFiltersFromLocation(locationSearch: H.Location['search']): CatalogComponentFilters {
    const parameters = new URLSearchParams(locationSearch)
    return {
        query: parameters.get('q') || undefined,
        owner: parameters.get('owner') || undefined,
        system: parameters.get('system') || undefined,
        tags: parameters.get('tags')?.split(','),
    }
}

// eslint-disable-next-line unicorn/prevent-abbreviations
function urlSearchParamsFromCatalogComponentFilters(
    filters: CatalogComponentFilters,
    base: URLSearchParams
): URLSearchParams {
    const parameters = new URLSearchParams(base)

    if (filters.query) {
        parameters.set('q', filters.query)
    } else {
        parameters.delete('q')
    }

    if (filters.owner) {
        parameters.set('owner', filters.owner)
    } else {
        parameters.delete('owner')
    }

    if (filters.system) {
        parameters.set('system', filters.system)
    } else {
        parameters.delete('system')
    }

    if (filters.tags) {
        parameters.set('tags', filters.tags.join(','))
    } else {
        parameters.delete('tags')
    }

    return parameters
}
