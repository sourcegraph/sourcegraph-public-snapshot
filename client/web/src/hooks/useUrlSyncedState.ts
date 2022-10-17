import { useState, useEffect } from 'react'

import { useHistory } from 'react-router'

/**
 * "useState" like hook that syncs state with URL search parameters.
 *
 * @param initialData initial data to use if no URL data is found
 */
export function useURLSyncedState<T extends Record<string, string>>(
    initialData: T,
    initialSearchParameters: URLSearchParams = new URLSearchParams(window.location.search),
    useHistoryHook = useHistory
): [T, (partialData: Partial<T>) => void] {
    const dataFromURL = Object.fromEntries(initialSearchParameters) as Partial<T>
    const [data, setData] = useState<T>({ ...initialData, ...dataFromURL })

    const setNewData = (partialData: Partial<T>): void => {
        setData(data => ({ ...data, ...partialData }))
    }

    const history = useHistoryHook()
    useEffect(() => {
        // Update the URL when the filters change
        const searchParameters = new URLSearchParams()
        for (const [key, value] of Object.entries(data)) {
            if (value !== undefined) {
                searchParameters.set(key, value)
            }
        }
        history?.replace({ search: searchParameters.toString() })
    }, [data, history])
    return [data, setNewData]
}
