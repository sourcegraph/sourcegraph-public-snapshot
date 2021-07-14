import { useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import { useDebounce } from '@sourcegraph/wildcard'

export const useQueryParameter = (key: string, fallback: string): [string, (value: string) => void] => {
    const location = useLocation()
    const history = useHistory()
    const searchParameters = useMemo(() => new URLSearchParams(location.search), [location.search])
    const [value, setValue] = useState(searchParameters.get(key) || fallback)
    const debouncedValue = useDebounce(value, 200)

    useEffect(() => {
        searchParameters.set(key, debouncedValue)
        const searchFragment = searchParameters.toString()
        if (location.search !== `?${searchFragment}`) {
            history.replace({
                search: searchFragment,
                hash: location.hash,
            })
        }
    }, [debouncedValue, history, key, location.hash, location.search, searchParameters])

    return [value, setValue]
}
