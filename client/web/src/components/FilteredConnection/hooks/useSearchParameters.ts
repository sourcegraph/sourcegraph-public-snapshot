import { useLocation } from 'react-router'

export const useSearchParameters = (): URLSearchParams => {
    const location = useLocation()
    return new URLSearchParams(location.search)
}
