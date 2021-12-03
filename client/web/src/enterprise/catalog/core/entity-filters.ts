import * as H from 'history'
import { useCallback, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

export interface CatalogEntityFilters {
    query: string
}

export interface CatalogEntityFiltersProps {
    filters: CatalogEntityFilters
    onFiltersChange: (newValue: CatalogEntityFilters) => void
}

export const useCatalogEntityFilters = (): CatalogEntityFiltersProps => {
    const history = useHistory()
    const location = useLocation()

    const filters = useMemo(() => filtersFromLocation(location.search), [location.search])
    const onFiltersChange = useCallback(
        (newValue: CatalogEntityFilters) => {
            history.push({
                ...location,
                search: urlSearchParametersFromFilters(newValue, new URLSearchParams(location.search)).toString(),
            })
        },
        [history, location]
    )

    return { filters, onFiltersChange }
}

function filtersFromLocation(locationSearch: H.Location['search']): CatalogEntityFilters {
    const parameters = new URLSearchParams(locationSearch)
    return {
        query: parameters.get('q') || '',
    }
}

function urlSearchParametersFromFilters(filters: CatalogEntityFilters, base: URLSearchParams): URLSearchParams {
    const parameters = new URLSearchParams(base)

    if (filters.query) {
        parameters.set('q', filters.query)
    } else {
        parameters.delete('q')
    }

    return parameters
}
