import * as H from 'history'
import { useCallback, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

export interface CatalogComponentFilters {
    query: string
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
        query: parameters.get('q') || '',
    }
}

// eslint-disable-next-line unicorn/prevent-abbreviations
function urlSearchParamsFromCatalogComponentFilters(
    filters: CatalogComponentFilters,
    base: URLSearchParams
): URLSearchParams {
    const parameters = new URLSearchParams(base)

    if (filters.query !== '') {
        parameters.set('q', filters.query)
    } else {
        parameters.delete('q')
    }

    return parameters
}
