import { useHistory, useLocation } from 'react-router'

import { getUrlQuery, GetUrlQueryParameters } from '../utils'

interface UseURLQueryParameters extends Omit<GetUrlQueryParameters, 'location'> {
    enabled?: boolean
}

export const useUrlQuery = ({ enabled, query, first, defaultFirst, visible }: UseURLQueryParameters): void => {
    const history = useHistory()
    const location = useLocation()

    if (!enabled) {
        return
    }

    const searchFragment = getUrlQuery({ query, first, defaultFirst, visible, location })

    if (searchFragment && location.search !== `?${searchFragment}`) {
        history.replace({
            search: searchFragment,
            hash: location.hash,
        })
    }
}
