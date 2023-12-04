import { useEffect } from 'react'

import { useNavigate, useLocation } from 'react-router-dom'

import { getUrlQuery, type GetUrlQueryParameters } from '../utils'

interface UseShowMorePaginationURLParameters extends Pick<GetUrlQueryParameters, 'first' | 'visibleResultCount'> {
    enabled?: boolean
}

/**
 * This hook replicates how FilteredConnection updates the URL when key variables change.
 * We use this to ensure the URL is kept in sync with the current connection state.
 * This is to allow users to build complex requests that can still be shared with others.
 * It is closely coupled to useShowMorePagination, which also derives initial state from the URL.
 */
export const useShowMorePaginationUrl = ({
    enabled,
    first,
    visibleResultCount,
}: UseShowMorePaginationURLParameters): void => {
    const location = useLocation()
    const navigate = useNavigate()
    const searchFragment = getUrlQuery({
        first,
        visibleResultCount,
        search: location.search,
    })

    useEffect(() => {
        if (enabled && searchFragment && location.search !== `?${searchFragment}`) {
            navigate(
                {
                    search: searchFragment,
                    hash: location.hash,
                },
                { replace: true }
            )
        }
    }, [enabled, navigate, location.hash, location.search, searchFragment])
}
