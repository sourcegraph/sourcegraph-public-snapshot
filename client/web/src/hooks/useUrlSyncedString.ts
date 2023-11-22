import { useSearchParams } from 'react-router-dom'

/**
 * "useState" like hook that syncs primitive string state with URL search parameters.
 *
 * @param defaultValue initial data to use if no URL search parameter is set
 */
export function useURLSyncedString(key: string, defaultValue: string): [string, (newValue: string) => void] {
    const [searchParams, setSearchParams] = useSearchParams()

    const setValue = (newValue: string): void => {
        setSearchParams(
            prevSearchParams => {
                const newSearchParams = new URLSearchParams(prevSearchParams)
                newSearchParams.set(key, newValue)
                return newSearchParams
            },
            { replace: true }
        )
    }

    return [searchParams.get(key) ?? defaultValue, setValue]
}
