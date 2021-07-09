import { Location } from 'history'
import { useHistory, useLocation } from 'react-router'

interface GetUrlQueryParameters {
    first?: number
    initialFirst?: number
    query?: string
    previousNodeCount?: number
    location: Location
}
const getUrlQuery = ({ query, first, initialFirst, previousNodeCount, location }: GetUrlQueryParameters): string => {
    const searchParameters = new URLSearchParams(location.search)

    if (query) {
        // TODO: Use FilteredConnection query_key
        searchParameters.set('query', query)
    }

    if (first !== initialFirst) {
        searchParameters.set('first', String(first))
    }

    // Eh? Visible === previousPage.length
    if (previousNodeCount && previousNodeCount !== 0) {
        searchParameters.set('visible', String(previousNodeCount))
    }

    return searchParameters.toString()
}

interface UseURLQueryParameters {
    enabled?: boolean
    first?: number
    initialFirst?: number
    query?: string
    // This is `visible` in original, TODO: check maps correctly
    previousNodeCount?: number
}

export const useUrlQuery = ({
    enabled,
    query,
    first,
    // TODO: Check support for initialFirst
    initialFirst,
    // TODO: Check support previousNodeCount
    previousNodeCount,
}: UseURLQueryParameters): void => {
    const history = useHistory()
    const location = useLocation()

    if (!enabled) {
        return
    }

    const searchFragment = getUrlQuery({ query, first, initialFirst, previousNodeCount, location })

    if (searchFragment && location.search !== `?${searchFragment}`) {
        history.replace({
            search: searchFragment,
            hash: location.hash,
        })
    }
}
