import { useCallback, useState, useEffect } from 'react'

import { useHistory } from 'react-router'

/**
 * "useState" like hook that syncs state with URL search parameters.
 *
 * @param defaultData initial data to use if no URL data is found
 */
export function useURLSyncedState<T extends Record<string, string>>(
    defaultData: T
): [T, (partialData: Partial<T>) => void] {
    const dataFromURL = Object.fromEntries(new URLSearchParams(location.search)) as Partial<T>
    const [data, setData] = useState<T>({ ...defaultData, ...dataFromURL })

    const setNewData = useCallback((partialData: Partial<T>) => {
        setData(data => ({ ...data, ...partialData }))
    }, [])

    const history = useHistory()
    useEffect(() => {
        // Update the URL when the filters change
        const searchParameters = new URLSearchParams()
        for (const [key, value] of Object.entries(data)) {
            if (value !== undefined) {
                searchParameters.set(key, value)
            }
        }
        history.replace({ search: searchParameters.toString() })
    }, [data, history])

    return [data, setNewData]
}
