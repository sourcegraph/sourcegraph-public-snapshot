import { useLocation } from 'react-router-dom'

/**
 * Return a new search parameters object based on the current URL.
 */
export const useSearchParameters = (): URLSearchParams => {
    const location = useLocation()
    return new URLSearchParams(location.search)
}
