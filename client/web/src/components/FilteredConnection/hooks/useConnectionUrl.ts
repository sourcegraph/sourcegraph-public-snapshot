import { useEffect } from 'react'
import { useHistory, useLocation } from 'react-router'

import { getUrlQuery, GetUrlQueryParameters } from '../utils'

interface UseConnectionURLParameters extends Pick<GetUrlQueryParameters, 'first' | 'visible'> {
    enabled?: boolean
}

/**
 * This hook replicates how FilteredConnection updates the URL when key variables change.
 */
export const useConnectionUrl = ({ enabled, first, visible }: UseConnectionURLParameters): void => {
    const location = useLocation()
    const history = useHistory()
    const searchFragment = getUrlQuery({
        first,
        visible,
        location,
    })

    useEffect(() => {
        if (enabled && searchFragment && location.search !== `?${searchFragment}`) {
            history.replace({
                search: searchFragment,
                hash: location.hash,
            })
        }
    }, [enabled, history, location.hash, location.search, searchFragment])
}
