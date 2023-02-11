import { useLocation } from 'react-router-dom-v5-compat'

/**
 * Return a new search parameters object based on the current URL.
 */
export const useSearchParameters = (): URLSearchParams => {
    const location = useLocation()
    return new URLSearchParams(location.search)
}
